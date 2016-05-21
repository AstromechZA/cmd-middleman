package main

import (
    "log"
    "os/exec"
    "os"
    "net"
    "net/rpc"
    "flag"
    "strings"
    "regexp"
    "fmt"
    "io/ioutil"

    "github.com/AstromechZA/middleman/transport"
)

const usageString =
`Middleman is a socket based RPC server and client for executing a set of
whitelisted commands on the other end of a unix file socket.

`

type MiddleManRPC struct {
    whitelistPatterns []string
}

func (self *MiddleManRPC) RunCmd(args *transport.CmdArgs, reply *transport.CmdResult) error {
    log.Printf("RunCmd being called with (%v %v)!", args.Cmd, args.Args)

    line := args.Cmd + " " + strings.Join(args.Args, " ")
    allowed := false
    for _, p := range self.whitelistPatterns {
        m, err := regexp.MatchString(p, line)
        if err != nil {
            log.Printf("Rexexp failed: %v", err.Error())
        }
        if m {
            allowed = true
        }
    }

    if allowed == false {
        reply.ReturnCode = 127
        reply.Output = "Command not found."
        return nil
    }

    cmd := exec.Command(args.Cmd, args.Args...)
    out, err := cmd.CombinedOutput()
    if err != nil {
        reply.ReturnCode = 1
    } else {
        reply.ReturnCode = 0
    }
    reply.Output = string(out)
    return nil
}


func main_inner() error {
    whitelistFlag := flag.String("whitelist", "", "Path to a file containing regex patterns. A regex pattern MUST match an incoming command for it to be run.")

    flag.Usage = func() {
        os.Stderr.WriteString(usageString)
        flag.PrintDefaults()
    }

    flag.Parse()

    // whitelist is required
    if *whitelistFlag == "" {
        return fmt.Errorf("Required --whitelist arg")
    }

    // extract patterns
    var whitelistPatterns []string
    {
        content, err := ioutil.ReadFile(*whitelistFlag)
        if err != nil { return err }
        whitelistPatterns = strings.Split(strings.TrimSpace(string(content)), "\n")
    }

    log.Println("server starting")
    socketPath := "/tmp/middleman.sock"
    os.Remove(socketPath)

    r := &MiddleManRPC{whitelistPatterns: whitelistPatterns}
    err := rpc.Register(r)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    addr := &net.UnixAddr{Net: "unix", Name: socketPath}
    listener, err := net.ListenUnix("unix", addr)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    for {
        conn, err := listener.AcceptUnix()
        log.Printf("Got connection %v", conn)
        if err != nil {
            log.Fatal(err)
        }
        rpc.ServeConn(conn)
    }
}


func main() {
    err := main_inner()
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
}
