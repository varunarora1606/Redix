package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/varunarora1606/My-Redis/rdb"
	"github.com/varunarora1606/My-Redis/resp"
	"github.com/varunarora1606/My-Redis/store"
)

var role string;
var master_replid string;
var master_repl_offset string;
var connected_slaves []net.Conn;
var master_host string;
var master_port string;
var port *string;
var currentMasterConn net.Conn;

func main() {

	dir := flag.String("dir", "./testdata", "Directory for RDB file")
	filename := flag.String("filename", "dump.rdb", "RDB file name")
	port = flag.String("port", "8000", "Port to run redis server")
	flag.Parse()

	ln, err := net.Listen("tcp", ":" + *port)
	if err != nil {
		panic(err)
	}

	role = "master"
	master_replid = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	master_repl_offset = "0"

	fmt.Println("Listening on :" + *port)

	if err := os.MkdirAll(*dir, os.ModePerm); err != nil {
		panic("Could not create directory: " + err.Error())
	}

	var kv store.Store = store.New()

	filepath := *dir + "/" + *filename

	if err := rdb.LoadRDB(filepath, kv); err != nil {
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

		go handleConn(conn, kv, filepath, "client")

		// TODO: SAVE just before shutdown

		// if err := rdb.SaveRDB(filepath, kv); err != nil {
		// 	panic("RDB file save error: " + err.Error())
		// }
		// fmt.Println("File saved successfully at " + filepath)
	}
}

func handleConn(conn net.Conn, kv store.Store, filepath string, connRole string) {
	defer conn.Close()
	defer fmt.Println("connection closed: " + connRole)

	for {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	fmt.Println(string(buf[:n]))

	msg, err := resp.Parse(string(buf[:n]))
	if err != nil {
		conn.Write([]byte("-" + err.Error() + "\r\n"))
		continue
	}

	fmt.Println("msg:", msg)

	switch msg[0] {
	case "PING":
		resp.WriteSimpleString(conn, "PONG")
	case "ECHO":
		if len(msg) < 2 {
			resp.WriteSimpleError(conn, "ERR wrong number of arguments for 'echo' command")
			// conn.Write([]byte("-ERR wrong number of arguments for 'echo' command\r\n"))
			continue
		}
		// conn.Write([]byte("$" + strconv.Itoa(len(msg[1])) + "\r\n" + msg[1] + "\r\n"))
		resp.WriteBulkString(conn, msg[1])
	case "SET":
		if role == "slave" && connRole == "client" {
			resp.WriteSimpleError(conn, "READONLY You can't write against a read only replica.")
			continue
		}
		if len(msg) < 3 {
			resp.WriteSimpleError(conn, "ERR wrong number of arguments for 'set' command")
			// conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
			continue
		}
		ttl := 0
		if len(msg) > 4 && msg[3] == "PX" {
			ttl, err = strconv.Atoi(msg[4])
			if err != nil {
				resp.WriteSimpleError(conn, "ERR expiry should be in int")
				// conn.Write([]byte("-ERR expiry should be in int\r\n"))
				continue
			}
		}
		kv.Set(msg[1], msg[2], int64(ttl))
		if role == "master" {
			for _, connected_slave := range connected_slaves {
				resp.WriteArray(connected_slave, msg)
			}
		}
		if connRole == "master" {
			continue
		}
		resp.WriteSimpleString(conn, "OK")
	case "GET":
		if len(msg) < 2 {
			resp.WriteSimpleError(conn, "ERR wrong number of arguments for 'get' command")
			continue
		}
		val, ok := kv.Get(msg[1])
		if !ok {
			resp.WriteBulkString(conn, "")
			continue
		}
		resp.WriteBulkString(conn, val)
	case "KEYS":
		if len(msg) < 2 {
			resp.WriteSimpleError(conn, "ERR wrong number of arguments for 'keys' command")
			// conn.Write([]byte("-ERR wrong number of arguments for 'keys' command\r\n"))
			continue
		}
		keys := kv.Keys(msg[1])
		resp.WriteArray(conn, keys)
	case "SAVE":
		if err := rdb.SaveRDB(filepath, kv); err != nil {
			resp.WriteSimpleError(conn, "ERR error while saving file: " + err.Error())
			// conn.Write([]byte("-ERR error while saving file: " + err.Error() + "\r\n"))
		}
		fmt.Println("File saved successfully at " + filepath)
		// conn.Write([]byte("+Ok\r\n"))
		resp.WriteSimpleString(conn, "OK")
	case "BGSAVE":
		go rdb.SaveRDB(filepath, kv)
		// conn.Write([]byte("+Background saving started\r\n"))
		resp.WriteSimpleString(conn, "Background saving started")
	case "INFO":
		if len(msg) < 2 && msg[1] != "replication" {
			resp.WriteSimpleError(conn, "ERR it should be `INFO replication`")
			continue
		}
		res := fmt.Sprintf("# Replication\r\nrole:%s\r\nconnected_slaves:%d\r\nmaster_replid:%s\r\nmaster_repl_offset:%s\r\n", role, len(connected_slaves), master_replid, master_repl_offset)
		resp.WriteBulkString(conn, res)
	case "REPLICAOF":
		if len(msg) < 3 {
			resp.WriteSimpleError(conn, "ERR wrong number of arguments for 'REPLICAOF' command")
			continue
		}
		if msg[1] == "NO" && msg[2] == "ONE" {
			role = "master"
			master_replid = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
			master_repl_offset = "0"
			continue
		} else {
			role = "slave"
			master_replid = "?"
			master_repl_offset = "-1"
			master_host = msg[1]
			master_port = msg[2]
			if _, err := strconv.Atoi(msg[2]); err != nil {
				resp.WriteSimpleError(conn, "ERR Invalid master port")
			}
		}
		go func() {
			masterConn, err := net.Dial("tcp", master_host + ":" + master_port)
			if err != nil {
				resp.WriteSimpleError(conn, "ERR Couldn't able to connect with master")
				return
			}

			if currentMasterConn != nil {
				currentMasterConn.Close()
				kv.FlushAll()
			}
			currentMasterConn = masterConn

			if err := doHandshake(masterConn, kv); err != nil {
				fmt.Println(err)
				resp.WriteSimpleError(conn, "ERR handshake failed")
				return
			}

			resp.WriteSimpleString(conn, "OK")
			handleConn(masterConn, kv, filepath, "master")
		}()
	case "REPLCONF":
		if len(msg) < 3 {
			resp.WriteSimpleError(conn, "ERR wrong number of arguments for 'REPLCONF' command")
			continue
		}
		if msg[1] == "listening-port" {
			resp.WriteSimpleString(conn, "OK")
		} else if msg[1] == "capa" {
			resp.WriteSimpleString(conn, "OK")
		} else {
			resp.WriteSimpleError(conn, fmt.Sprintf("ERR Unrecognized REPLCONF option: %s", msg[2]))
		}
	case "PSYNC":
		if len(msg) < 3 {
			resp.WriteSimpleError(conn, "ERR wrong number of arguments for 'PSYNC' command")
			continue
		}
		if role == "slave" {
			resp.WriteSimpleError(conn, "ERR target instance is not a master")
			continue
		}
		if msg[1] == "?" && msg[2] == "-1" {
			resp.WriteSimpleString(conn, fmt.Sprintf("FULLRESYNC %s %s", master_replid, master_repl_offset))
			rdb.SaveRDB(filepath, kv)
			err := sendRdbFile(conn, filepath)
			if err != nil {
				resp.WriteSimpleError(conn, "ERR rdb file transer failed")
				continue
			}
		}
		fmt.Println("HandShake Completed")
		connected_slaves = append(connected_slaves, conn)
	default:
		if connRole == "master" {
			continue
		}
		resp.WriteSimpleError(conn, "ERR unknown command")
	}
	}

}


func sendRdbFile(conn net.Conn,filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("No data. PSYNC error: " + err.Error())
		return err
	}
	defer file.Close()
	if err := resp.WriteRDB(conn, file); err != nil {
		return err
	}
	fmt.Println("RDB sent successfully")
	return nil
}


func readLine(conn net.Conn) (string, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), err
}

func doHandshake(conn net.Conn, kv store.Store) error {
    // PING
    resp.WriteArray(conn, []string{"PING"})
    pong, err := readLine(conn)
    if err != nil || pong != "+PONG\r\n" {
        return fmt.Errorf("invalid PONG: %s", pong)
    }

    // REPLCONF 1
    resp.WriteArray(conn, []string{"REPLCONF", "listening-port", *port})
    ok1, err := readLine(conn)
    if err != nil || ok1 != "+OK\r\n" {
        return fmt.Errorf("invalid REPLCONF OK1: %s", ok1)
    }

    // REPLCONF 2
    resp.WriteArray(conn, []string{"REPLCONF", "capa", "psync2"})
    ok2, err := readLine(conn)
    if err != nil || ok2 != "+OK\r\n" {
        return fmt.Errorf("invalid REPLCONF OK2: %s", ok2)
    }
	fmt.Println("2nd handshake completed")

    // PSYNC
    resp.WriteArray(conn, []string{"PSYNC", master_replid, master_repl_offset})
	reader := bufio.NewReader(conn)
    fullresync, err := reader.ReadString('\n')
    if err != nil {
		fmt.Println("handshake error")
        return fmt.Errorf("invalid FULLRESYNC: %v", err)
    }
	fmt.Println("3rd handshake completed")
    fmt.Println("FULLRESYNC response:", fullresync)
	fullresyncArr := strings.Split(fullresync, " ")
	master_replid = fullresyncArr[1]
	master_repl_offset = fullresyncArr[2]


    rdbSize, _ := reader.ReadString('\n')
	fmt.Println("RDB file size:", rdbSize)
	if err = rdb.LoadRDBFromReader(reader, kv); err != nil {
		return err
	}
    return nil
}

// TODO: Add the saving of rdb into in memory store.
// TODO: In loadfile if file do not exist then create file.