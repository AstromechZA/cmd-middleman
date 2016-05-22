package main

import (
    "flag"
    "fmt"
    "net/rpc"
    "os"

    "github.com/AstromechZA/cmd-middleman/transport"
    "github.com/AstromechZA/cmd-middleman/common"
)

const usageString =
`rpc-cmd-client is the client side of an RPC system for executing commands
across a unix domain socket. This can be used for escalating priviledges of the
caller for specific commands or calling commands outside of a Docker container.

This has a high chance of being used maliciously if it results in privilege
escalation, but is certainly better than some alternatives.

The exit code will be 120 if an error regarding the local client occurs. Any
other exit codes will be those of the RPC call.

Arguments will be passed through to the server side for execution. For example:
$ rpc-cmd-client --socket /tmp/s.sock ls -al
Will execute 'ls' with arguments '-al' on the server side.

If the command being executed does not match a whitelist pattern on the server,
it will exit with code 127 'Command not found.'.
`

const rpcCall = "MiddleManRPC.RunCmd"

// Version is the version string
// format should be 'X.YZ'
// Set this at build time using the -ldflags="-X common.Version=X.YZ"
var Version = "<unofficial build>"

func mainInner() error {
    // command line flags
    socketFileFlag := flag.String("socket", "", "Path to a unix socket file being listened to by the server.")
    versionFlag := flag.Bool("version", false, "Print the version string")

    // set up the usage strings
    flag.Usage = func() {
        fmt.Println(usageString)
        flag.PrintDefaults()
    }

    // parse the args
    flag.Parse()

    // print version if required
    if *versionFlag {
        fmt.Println(Version)
        os.Exit(0)
    }

    // required args
    if err := common.RequiredFlag("socket"); err != nil {
        return common.UsageError(usageString, err)
    }

    // security checks on the file
    st, err := os.Stat(*socketFileFlag)
    // make sure it exists on the system
    if os.IsNotExist(err) { return err }
    // make sure it is a unix socket
    if (st.Mode() & os.ModeSocket) != os.ModeSocket {
        return fmt.Errorf("Socket file is not a unix socket")
    }
    // make sure its permissions are 0o600
    if (st.Mode() & 0777) != 0600 {
        return fmt.Errorf("Expected socket file permissions to be 0600")
    }

    // now we parse the remaining arguments as the actual command
    if flag.NArg() == 0 {
        return common.UsageError(usageString, fmt.Errorf("Required at least one argument"))
    }
    // the first one is the program
    rpcProgram := flag.Args()[0]
    // the rest are any args for it
    rpcArgs := flag.Args()[1:]

    // now lets connect to this socket file
    client, err := rpc.Dial("unix", *socketFileFlag)
    if err != nil { return err }

    args := &transport.CmdArgs{Cmd: rpcProgram, Args: rpcArgs}
    reply := &transport.CmdResult{ReturnCode: -1}
    if err = client.Call(rpcCall, args, reply); err != nil { return err }
    if err = client.Close(); err != nil { return err }

    // print the combined output
    fmt.Print(reply.Output)
    // exit with the return code
    os.Exit(reply.ReturnCode)

    // no error
    return nil
}

func main() {
    if err := mainInner(); err != nil {
        os.Stderr.WriteString(err.Error() + "\n")
        // we use exit 120 here because exit 1 would make it too confusing as
        // to which program failed: remote or local.
        os.Exit(120)
    }
}
