package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"fmt"

	"strings"

	"strconv"

	"github.com/vishvananda/netns"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("%s <pid>", os.Args[0])
		return
	}
	var namespacePid int
	if os.Args[1] == "subprocess" {
		if len(os.Args) > 2 {
			namespacePid, _ = strconv.Atoi(os.Args[2])
		}
		subProcess(namespacePid)
		return
	}

	namespacePid, _ = strconv.Atoi(os.Args[1])
	cmd := exec.Command(os.Args[0], "subprocess", strconv.Itoa(namespacePid))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Start()

	log.Printf("Start Network Namespace test PID=%d SubPID=%d NSPID\n", os.Getpid(), cmd.Process.Pid, namespacePid)

	go bufio.NewReader(stdout).WriteTo(os.Stdout)

	if namespacePid == 0 {
		namespacePid = cmd.Process.Pid
	}
	defaultns, err := netns.Get()
	log.Printf("Current Network Namespace (pid=%d) (%s) (err=%s)\n", os.Getpid(), defaultns.String(), err)
	nsh, err := netns.GetFromPid(os.Getpid())

	log.Print("[Change ns] Type Enter")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	nsh, err = netns.GetFromPid(namespacePid)
	netns.Set(nsh)
	log.Printf("Current Network Namespace (pid=%d) (%s) (err=%s)\n", namespacePid, nsh.String(), err)

	log.Print("[TCP Connect] Type Enter")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	defer conn.Close()
	if err != nil {
		log.Println("Dial error", err)
		os.Exit(1)
	}
	log.Print("[Return to default NNS] Type Enter")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	nsh.Close()
	netns.Set(defaultns)

	time.Sleep(10000000)
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}

	for {
		time.Sleep(100000000)
	}
}

func subProcess(pid int) {

	var log func(format string, v ...interface{}) = func(format string, v ...interface{}) {
		logMessage := fmt.Sprintf(format, v...)
		fmt.Printf("    [%d]%s", os.Getpid(), logMessage)
	}
	log("Start Network NS Sub Proccess PID=%d PPID=%d NSID=%d\n", os.Getpid(), os.Getppid(), pid)

	var ns netns.NsHandle
	var err error
	if pid == 0 {
		// Create a new network namespace
		ns, err = netns.New()

	} else {
		ns, err = netns.GetFromPid(pid)
	}
	netns.Set(ns)
	log("Network Namespace PID=%d is %s: err=%s\n", pid, ns.String(), err)
	defer ns.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log("Listen error %s\n", err)
		return
	}
	defer ln.Close()
	log("Listen now on %v\n", ln.Addr())

	// accept connection on port
	conn, err := ln.Accept()
	if err != nil {
		log("Accept error %s\n", err)
		return
	}
	defer conn.Close()
	log("Connection established From %v\n", conn.RemoteAddr())

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// output message received
		log("Message Received: %s\n", string(message))
		// sample process for string received
		newmessage := strings.ToUpper(message)
		// send new string back to client
		conn.Write([]byte(newmessage + "\n"))
	}
}
