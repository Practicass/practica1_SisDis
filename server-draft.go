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
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"practica1/com"
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
	
}
