package main

//func main() {
//	arguments := os.Args
//	if len(arguments) == 1 {
//		fmt.Println("Please provide a host:port string")
//		return
//	}
//	CONNECT := arguments[1]
//
//	s, err := net.ResolveUDPAddr("udp4", CONNECT)
//	c, err := net.DialUDP("udp4", nil, s)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
//	defer c.Close()
//
//	data := []byte(string("heartbeat") + "\n")
//	_, err = c.Write(data)
//
//	for {
//		buffer := make([]byte, 8192)
//		n, _, err := c.ReadFromUDP(buffer)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//
//		if strings.TrimSpace(string(buffer[0:n])) == "STOP" {
//			fmt.Println("Exiting UDP client!")
//			return
//		}
//
//		fmt.Println("Received cmd: " + string(buffer[0:n]))
//		cmdWithArgs := strings.Split(string(buffer[0:n]), " ")
//		command, args := cmdWithArgs[0], cmdWithArgs[1:]
//
//		cmd := exec.Command(command, args...)
//		stdout, err := cmd.Output()
//
//		if err != nil {
//			fmt.Println(err.Error())
//			stdout = []byte("ERROR: " + err.Error() + "\n")
//		}
//
//		data = []byte(string(stdout) + "\n")
//		_, err = c.Write(data)
//
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//	}
//}
