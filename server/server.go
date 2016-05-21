package main

import (
    "bytes"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "net/rpc"
    "os"
    "os/exec"
    "os/signal"
    "regexp"
    "strings"
    "syscall"

    "github.com/AstromechZA/middleman/transport"
    "github.com/AstromechZA/middleman/common"
)

const usageString =
`rpc-cmd-server is the server side of an RPC system for executing commands
across a unix domain socket. This can be used for escalating priviledges of the
caller for specific commands or calling commands outside of a Docker container.

This has a high chance of being used maliciously if it results in privilege
escalation, but is certainly better than some alternatives.

To protect against running any old command this application requires a file
containing a number of regex patterns for whitelisted commands. At least one
of these patterns MUST match a command and its arguments in order to be executed.
`

// MiddleManRPC is an rpc identifier whose only attribute is a list of
// whitelisting regex patterns.
type MiddleManRPC struct {
    whitelistPatterns []regexp.Regexp
}

// RunCmd call allows an rpc call to execute a process on this host
func (s *MiddleManRPC) RunCmd(args *transport.CmdArgs, reply *transport.CmdResult) error {
    log.Printf("RunCmd being called with '%v %v'", args.Cmd, args.Args)

    line := strings.TrimSpace(args.Cmd + " " + strings.Join(args.Args, " "))
    allowed := false
    for i, p := range s.whitelistPatterns {
        lineBytes := []byte(line)
        result := p.Find(lineBytes)
        if result != nil && bytes.Equal(result, lineBytes) {
            log.Printf("Command matched pattern %v: '%v'", i, p.String())
            allowed = true
            break
        }
    }

    if allowed == false {
        log.Printf("Command did not match any of the whitelist patterns")
        reply.ReturnCode = 127
        reply.Output = "Command not found."
        return nil
    }

    log.Printf("Executing command '%v %v'", args.Cmd, args.Args)
    cmd := exec.Command(args.Cmd, args.Args...)
    // we could possibly change this to separate out stdout and stderr
    out, err := cmd.CombinedOutput()
    if err == nil {
        reply.Output = string(out)
        reply.ReturnCode = 0
    } else {
        reply.ReturnCode = 1
        if ee, ok := err.(*exec.ExitError); ok {
            reply.Output = string(out)
            if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
                reply.ReturnCode = ws.ExitStatus()
            }
        } else {
            reply.Output = err.Error()
        }
    }
    log.Printf("Result was %d bytes of output with return code %d", len(out), reply.ReturnCode)
    return nil
}

func compilePatternsFromFile(path string) (*[]regexp.Regexp, error) {
    content, err := ioutil.ReadFile(path)
    if err != nil { return nil, err }
    whitelistPatterns := strings.Split(strings.TrimSpace(string(content)), "\n")
    whitelistRegexes := make([]regexp.Regexp, len(whitelistPatterns))
    for i, p := range whitelistPatterns {
        r, err := regexp.Compile(p)
        if err != nil { return nil, err }
        whitelistRegexes[i] = *r
    }
    return &whitelistRegexes, nil
}

func mainInner() error {
    // command line flags
    whitelistFlag := flag.String("whitelist", "",
        "Path to a file containing regex patterns. A regex pattern MUST match an incoming command for it to be run.")
    socketFileFlag := flag.String("socket", "",
        "Path to a unix socket file being listened to by the server.")

    // usage string
    flag.Usage = func() {
        os.Stderr.WriteString(usageString)
        flag.PrintDefaults()
    }

    // parse the args
    flag.Parse()

    // required args
    for _, n := range []string{"whitelist", "socket"} {
        if err := common.RequiredFlag(n); err != nil {
            return common.UsageError(usageString, err)
        }
    }

    // extract and compile patterns
    whitelistPatterns, err := compilePatternsFromFile(*whitelistFlag)
    if err != nil { return err }

    // validate socket things
    log.Printf("Checking socket file %v", *socketFileFlag)
    st, err := os.Stat(*socketFileFlag)
    // we dont really care if it doesn't exit
    // we want to find out whether we can safely delete it
    if os.IsNotExist(err) == false {
        // make sure it is a unix socket
        if (st.Mode() & os.ModeSocket) != os.ModeSocket {
            return fmt.Errorf("Socket file %v exists but is not a unix socket. Please delete it for me.")
        }
        // otherwise we can delete it
        log.Printf("Deleting old socket file %v", *socketFileFlag)
        if err := os.Remove(*socketFileFlag); err != nil {
            return err
        }
    }

    // now register rpc interface object
    log.Println("Registering RPC interface")
    r := &MiddleManRPC{whitelistPatterns: *whitelistPatterns}
    if err := rpc.Register(r); err != nil { return err }

    // build socket address
    addr := &net.UnixAddr{Net: "unix", Name: *socketFileFlag}
    // listen
    log.Printf("Listening on socket %v", addr)
    listener, err := net.ListenUnix("unix", addr)
    if err != nil { return err }
    // chmod it to 0600
    log.Println("Chmodding socket file to 0600")
    if err := os.Chmod(*socketFileFlag, 0600); err != nil { return err }
    // make sure we remove it at the end
    defer os.Remove(*socketFileFlag)

    stopped := false
    log.Println("Beginning connection loop..")
    go func() {
        for stopped == false {
            conn, err := listener.AcceptUnix()
            if err != nil {
                log.Fatal(err)
            }
            rpc.ServeConn(conn)
        }
    }()

    // instead of sitting in a for loop or something, we wait for sigint
    signalChannel := make(chan os.Signal, 1)
    // notify that we are going to handle interrupts
    signal.Notify(signalChannel, os.Interrupt)
    for sig := range signalChannel {
        log.Printf("Received %v signal. Stopping.", sig)
        stopped = true
        if err := listener.Close(); err != nil { return err }
        return nil
    }
    return nil
}

func main() {
    if err := mainInner(); err != nil {
        os.Stderr.WriteString(err.Error() + "\n")
        os.Exit(1)
    }
}
