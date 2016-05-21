package transport

type CmdArgs struct {
    Cmd string
    Args []string
}

type CmdResult struct {
    ReturnCode int
    Output string
}
