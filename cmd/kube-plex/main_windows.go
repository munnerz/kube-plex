package main

import "os"

// Only signal to catch in windows is os.Interrupt
var shutdownSignals = []os.Signal{os.Interrupt}

// protectSigKill is a no-op on Windows
func protectSigKill(ctx context.Context) context.Context { return ctx }
