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
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"practica1/com"
	"strconv"

	"bufio"


	"golang.org/x/crypto/ssh"
)

type requestEncoder struct {
	req     com.Request
	encoder *gob.Encoder
}

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

func terceraArq(requestChan chan requestEncoder) {
	for {
		request := <-requestChan
		primos := FindPrimes(request.req.Interval)
		reply := com.Reply{request.req.Id, primos}
		err := request.encoder.Encode(reply)
		checkError(err)

	}
}

func cuartaArq(requestChan chan requestEncoder, ip string) {

	for {

		sshConfig := &ssh.ClientConfig{
			User: "a842255",
			Auth: []ssh.AuthMethod{
				// You can use password or key authentication here.
				// For key authentication, load your private key.
				// Example:
				// ssh.PublicKeys(privateKey),
				ssh.Password("philha32"),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: Insecure for production use
		}

		client, err := ssh.Dial("tcp", ip, sshConfig)
		if err != nil {
			log.Fatalf("Failed to dial: %s", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			log.Fatalf("Failed to create session: %s", err)
		}
		defer session.Close()

		requestEnc := <-requestChan
		request := requestEnc.req

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
    	       me(i) {
    	               primfor i := 2; (i < n) && !foundDivisor; i++ {
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
    	           if IsPries = append(primes, i)
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

		var stdout, stderr bytes.Buffer
		session.Stdout = &stdout
		session.Stderr = &stderr



		err = session.Run("echo '" + functionCode + "' > custom_function.go && go run custom_function.go " + strconv.Itoa(request.Id) + " " + strconv.Itoa(request.Interval.A) + " " + strconv.Itoa(request.Interval.B))
		if err != nil {
			log.Fatalf("Failed to run custom function: %s", err)
		}

		// Retrieve the output
		output := []byte(stdout.String())
		//fmt.Println(string(output))
		var reply com.Reply

		err = json.Unmarshal(output, &reply)
		if err != nil {
			log.Fatalf("Failed in json")
		}

		err = requestEnc.encoder.Encode(&reply)
		checkError(err)

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
	//requestChan := make(chan requestEncoder)
	//for i:=0; i<6; i++ {
	//	go terceraArq(requestChan)
	//}
	//for{
	//	conn, err := listener.Accept()
	//
	//	checkError(err)
	//	var requestEnc requestEncoder
	//	requestEnc.encoder = gob.NewEncoder(conn)
	//	decoder := gob.NewDecoder(conn)
	//
	//	err = decoder.Decode(&requestEnc.req)
	//	checkError(err)
	//
	//
	//	requestChan <- requestEnc
	//
	//
	//}
	//---------------------------------------------------------------------------
	requestChan := make(chan requestEncoder)
	file, err := os.Open("file.txt")
	checkError(err)
	fileScanner := bufio.NewScanner(file)
	for i := 0; i < 4; i++ {
		fileScanner.Scan()
		ip := fileScanner.Text()
		go cuartaArq(requestChan, ip)
	}
	for {
		conn, err := listener.Accept()

		checkError(err)
		var requestEnc requestEncoder
		requestEnc.encoder = gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)

		err = decoder.Decode(&requestEnc.req)
		checkError(err)

		requestChan <- requestEnc
	}

}
