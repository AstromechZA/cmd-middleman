package common

import (
    "flag"
    "fmt"
)

// UsageError take an error and before returning it, prints the usage string
// and flag arguments description.
func UsageError(usageString string, e error) error {
    fmt.Println(usageString)
    flag.PrintDefaults()
    fmt.Println()
    return e
}

// RequiredFlag is a little bit of a silly method, an experiment in checking
// required args
func RequiredFlag(name string) error {
    f := flag.Lookup(name)
    if f == nil { return fmt.Errorf("No flag named %v", name) }
    if f.Value.(flag.Getter).Get() == f.DefValue {
        return fmt.Errorf("--%v is a required argument", f.Name)
    }
    return nil
}
