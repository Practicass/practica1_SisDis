/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes a la práctica 1
 */
 package main

 import (
	 "bytes"
	 "encoding/gob"
	 "fmt"
	 "log"
	 "net"
	 "os"
	 "practica1/com"
	 "strconv"
	 "encoding/json"
 
	 "golang.org/x/crypto/ssh"
 )
 
 func checkError(err error) {
	 if err != nil {
		 fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		 os.Exit(1)
	 }
 }
 
 // PRE: verdad
 // POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
 func IsPrime(n int) (foundDivisor bool) {
	 foundDivisor = false
	 for i := 2; (i < n) && !foundDivisor; i++ {
		 foundDivisor = (n%i == 0)
	 }
	 return !foundDivisor
 }
 
 // PRE: interval.A < interval.B
 // POST: FindPrimes devuelve todos los números primos comprendidos en el
 //
 //	intervalo [interval.A, interval.B]
 func FindPrimes(interval com.TPInterval) (primes []int) {
	 for i := interval.A; i <= interval.B; i++ {
		 if IsPrime(i) {
			 primes = append(primes, i)
		 }
	 }
	 return primes
 }
 
 func receiveMessage(CONN_HOST string, CONN_PORT string, conn net.Conn) {
	 var request com.Request
 
	 encoder := gob.NewEncoder(conn)
	 decoder := gob.NewDecoder(conn)
	 err := decoder.Decode(&request)
	 checkError(err)
 
	 primos := FindPrimes(request.Interval)
	 reply := com.Reply{request.Id, primos}
	 err = encoder.Encode(reply)
	 checkError(err)
 
	 conn.Close()
 }
 
 func terceraArq(requestChan chan com.Request, replyChan chan com.Reply) {
	 for {
		 request := <-requestChan
		 primos := FindPrimes(request.Interval)
		 reply := com.Reply{request.Id, primos}
		 replyChan <- reply
 
	 }
 }
 
 func cuartaArq(requestChan chan com.Request, replyChan chan com.Reply) {
	 for {
		 // SSH client configuration
		 sshConfig := &ssh.ClientConfig{
			 User: "as",
			 Auth: []ssh.AuthMethod{
				 // You can use password or key authentication here.
				 // For key authentication, load your private key.
				 // Example:
				 // ssh.PublicKeys(privateKey),
				 ssh.Password("as"),
			 },
			 HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: Insecure for production use
		 }
 
		 // Connect to the remote server
		 client, err := ssh.Dial("tcp", "192.168.56.2:22", sshConfig)
		 if err != nil {
			 log.Fatalf("Failed to dial: %s", err)
		 }
		 defer client.Close()
 
		 // Create a session
		 session, err := client.NewSession()
		 if err != nil {
			 log.Fatalf("Failed to create session: %s", err)
		 }
		 defer session.Close()
 
		 request := <-requestChan
		 fmt.Println(request)
 
		 // Define a custom function in Go source code
		 functionCode := fmt.Sprint(`
			 package main
 
			 import (
				 "fmt"
				 "os"
				 "strconv"
				 "encoding/json"
			 )
 
			 type TPInterval struct {
				 A int
				 B int
			 }
			 
			 type Request struct {
				 Id int
				 Interval TPInterval
			 }
 
			 type Reply struct {
				 Id int
				 Primes []int
			 }
			 
 
			 // PRE: verdad
			 // POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
			 func IsPrime(n int) (foundDivisor bool) {
				foundDivisor = false
				for i := 2; (i < n) && !foundDivisor; i++ {
					foundDivisor = (n%i == 0)
				}
				return !foundDivisor
			}
		
			// PRE: interval.A < interval.B
			// POST: FindPrimes devuelve todos los números primos comprendidos en el
			//
			//	intervalo [interval.A, interval.B]
			func FindPrimes(interval TPInterval) (primes []int) {
				for i := interval.A; i <= interval.B; i++ {
					if IsPrime(i) {
						primes = append(primes, i)
					}
				}
				return primes
			}
 
			 func main() {
				 id,_:=strconv.Atoi(os.Args[1])
				 a,_:=strconv.Atoi(os.Args[2])
				 b,_:=strconv.Atoi(os.Args[3])
				 intervalo := TPInterval{a,b}
				 request := Request{id, intervalo}
				 primos := FindPrimes(request.Interval)
				 reply := Reply{request.Id, primos}
				 jsonReply, _ := json.Marshal(reply)
				 fmt.Println(string(jsonReply))
			 }
 
		 `)
 
		 // Execute the function code remotely and capture the output
		 var stdout, stderr bytes.Buffer
		 session.Stdout = &stdout
		 session.Stderr = &stderr
 
		 // Save the function code to a temporary file and run it
		 err = session.Run("echo '" + functionCode + "' > custom_function.go && go run custom_function.go " + strconv.Itoa(request.Id) + " " + strconv.Itoa(request.Interval.A) + " " + strconv.Itoa(request.Interval.B))
		 if err != nil {
			 log.Fatalf("Failed to run custom function: %s", err)
		 }
 
		 // Retrieve the output
		 output := []byte(stdout.String())
 
		 var reply com.Reply
 
		 err = json.Unmarshal(output, &reply)
		 if err != nil {
			 log.Fatalf("Failed in json")
		 }
 
		 // Print or return the output as needed
		 replyChan <- reply
	 }
 }
 
 func main() {
	 var CONN_TYPE = "tcp"
	 var CONN_HOST = "127.0.0.1"
	 var CONN_PORT = "30000"
	 listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	 checkError(err)
	 //--------------------------------------------------------------
	 //-----------------------PRIMERA ARQUITECTURA-------------------
	 //--------------------------------------------------------------
	 //for {
	 //	conn, err := listener.Accept()
	 //	defer conn.Close() //se ejecutara al final, es añadido a la pila
	 //	checkError(err)
	 //
	 //	receiveMessage(CONN_HOST, CONN_PORT, conn)
	 //}
 
	 //--------------------------------------------------------------
	 //-----------------------SEGUNDA ARQUITECTURA-------------------
	 // //--------------------------------------------------------------
	 //for {
	 //	conn, err := listener.Accept()
	 //	defer conn.Close() //se ejecutara al final, es añadido a la pila
	 //	checkError(err)
	 //
	 //	go receiveMessage(CONN_HOST, CONN_PORT, conn)
	 //}
 
	 //--------------------------------------------------------------
	 //-----------------------TERCERA ARQUITECTURA-------------------
	 // //--------------------------------------------------------------
	 requestChan := make(chan com.Request)
	 replyChan := make(chan com.Reply)
	 for i := 0; i < 4; i++ {
		 go cuartaArq(requestChan, replyChan)
	 }
	 for {
		 conn, err := listener.Accept()
 
		 checkError(err)
 
		 var request com.Request
 
		 encoder := gob.NewEncoder(conn)
		 decoder := gob.NewDecoder(conn)
		 err = decoder.Decode(&request)
		 checkError(err)
 
		 requestChan <- request
 
		 reply := <-replyChan
		 err = encoder.Encode(reply)
		 checkError(err)
 
		 conn.Close()
 
	 }
 }
 