package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/varunarora1606/My-Redis/rdb"
	"github.com/varunarora1606/My-Redis/resp"
	"github.com/varunarora1606/My-Redis/store"
)

func main() {
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening on :8000")

	dir := flag.String("dir", "./testdata", "Directory for RDB file")
	filename := flag.String("filename", "dump.rdb", "RDB file name")
	flag.Parse()

	if err := os.MkdirAll(*dir, os.ModePerm); err != nil {
		panic("Could not create directory: " + err.Error())
	}

	var kv store.Store = store.New()

	filepath := *dir + "/" + *filename
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fmt.Println("RDB file not found. Starting fresh.")
	} else if err := rdb.LoadRDB(filepath, kv); err != nil {
		panic("RDB file error: " + err.Error())
	}
	fmt.Println("Data loaded successfully from " + filepath)

	for {
		fmt.Println("hello")
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		fmt.Println("Connection established")

		go handleClient(conn, kv, filepath)

		// TODO:
		// time.Sleep(10 * time.Second)

		// if err := rdb.SaveRDB(filepath, kv.SnapShot()); err != nil {
		// 	panic("RDB file save error: " + err.Error())
		// }
		// fmt.Println("File saved successfully at " + filepath)
	}
}

func handleClient(conn net.Conn, kv store.Store, filepath string) {
	defer conn.Close()
	for {
		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection error:", err)
			return
		}

		msg, err := resp.Parse(string(buf[:n]))
		if err != nil {
			conn.Write([]byte("-" + err.Error() + "\r\n"))
			continue
		}

		fmt.Println("msg:", msg)

		switch msg[0] {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(msg) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'echo' command\r\n"))
				continue
			}
			conn.Write([]byte("$" + strconv.Itoa(len(msg[1])) + "\r\n" + msg[1] + "\r\n"))
		case "SET":
			if len(msg) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				continue
			}
			ttl := 0
			if len(msg) > 4 && msg[3] == "PX" {
				ttl, err = strconv.Atoi(msg[4])
				if err != nil {
					conn.Write([]byte("-ERR expiry should be in int\r\n"))
					continue
				}
			}
			kv.Set(msg[1], msg[2], int64(ttl))
			conn.Write([]byte("+Ok\r\n"))
		case "GET":
			if len(msg) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue
			}
			val, ok := kv.Get(msg[1])
			if !ok {
				conn.Write([]byte("$" + strconv.Itoa(-1) + "\r\n"))
				continue
			}
			conn.Write([]byte("$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"))
		case "KEYS":
			if len(msg) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'keys' command\r\n"))
				continue
			}
			keys := kv.Keys(msg[1])
			if len(keys) == 0 {
				conn.Write([]byte("$" + strconv.Itoa(-1) + "\r\n"))
				continue
			}
			res := "*" + strconv.Itoa(len(keys)) + "\r\n"
			for _, key := range keys {
				res = res + "$" + strconv.Itoa(len(key)) + "\r\n" + key + "\r\n"
			}
			conn.Write([]byte(res))
		case "SAVE":
			if err := rdb.SaveRDB(filepath, kv.SnapShot()); err != nil {
				conn.Write([]byte("-ERR error while saving file: " + err.Error() + "\r\n"))
			}
			fmt.Println("File saved successfully at " + filepath)
			conn.Write([]byte("+Ok\r\n"))
		case "BGSAVE":
			go rdb.SaveRDB(filepath, kv.SnapShot())
			conn.Write([]byte("+Background saving started\r\n"))
		default:
			conn.Write([]byte("-ERR unknown command\r\n"))
		}
	}

}
