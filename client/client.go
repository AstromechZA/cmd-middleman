package main

import (
    "log"
    "net/rpc"
    "os"
    "flag"
    "fmt"

    "github.com/AstromechZA/middleman/transport"
)

func main_inner() error {
    socketFileFlag := flag.String("socket", "", "Path to a unix socket file being listened to by the server.")

    flag.Parse()

    // whitelist is required
    if *socketFileFlag == "" {
        return fmt.Errorf("Required --socket arg")
    }
    // make sure it exists as a file on the system
    if _, err := os.Stat(*socketFileFlag); os.IsNotExist(err) {
        return err
    }

    if len(flag.Args()) == 0 {
        return fmt.Errorf("Required at least one cmd argument")
    }

    cmd0 := flag.Args()[0]
    cmd1 := flag.Args()[1:]

    log.Println("client starting")
    client, err := rpc.Dial("unix", *socketFileFlag)
    if err != nil { return err }
    log.Println("connection established")

    args := &transport.CmdArgs{Cmd: cmd0, Args: cmd1}
    //args := &transport.CmdArgs{Cmd: "ls", Args: []string{"-al", "/"}}
    reply := &transport.CmdResult{ReturnCode: -1}
    err = client.Call("MiddleManRPC.RunCmd", args, reply)
    if err != nil {
        log.Fatal(err)
    }
    log.Println(reply)

    log.Println("calling client.Close()")
    if err := client.Close(); err != nil {
        log.Println("client.Close() error: ", err)
    }

    log.Println("client exiting")

    return nil
}

func main() {
    err := main_inner()
    if err != nil {
        os.Stderr.WriteString(err.Error() + "\n")
        os.Exit(1)
    }
}
