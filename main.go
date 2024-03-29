package main

import (
  "bufio"
  "crypto/sha256"
  "encoding/hex"
  "encoding/json"
  "io"
  "log"
  "net"
  "os"
  "strconv"
  "time"

  "github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

//Block represents each 'item' in the blockchain
type Block struct {
  Index     int
  Timestamp string
  BPM       int
  Hash      string
  PrevHash  string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block

// SHA256 hashing function
func calculateHash(block Block) string {
  record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
  h := sha256.New()
  h.Write([]byte(record))
  hashed := h.Sum(nil)
  return hex.EncodeToString(hashed)
}

// Create a new block using previus blocks hashed
func generateBlock(oldBlock Block, BPM int) (Block, error) {
  var newBlock Block

  t := time.Now()

  newBlock.Index = oldBlock.Index + 1
  newBlock.Timestamp = t.String()
  newBlock.BPM = BPM
  newBlock.PrevHash = oldBlock.Hash
  newBlock.Hash = calculateHash(newBlock)

  return newBlock, nil
}

// Make sure the block is valid by checking index and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
  if oldBlock.Index+1 != newBlock.Index {
    return false
  }

  if oldBlock.Hash != newBlock.PrevHash {
    return false
  }

  if calculateHash(newBlock) != newBlock.Hash {
    return false
  }

  return true
}

//Make sure the chain we are checking is longer than the current blockchain
func replaceChain(newBlocks []Block) {
  if len(newBlocks) > len(Blockchain) {
    Blockchain = newBlocks
  }
}

//Until here is all of the last projetct (blockchain) with the http stuff stripped out.

//bcServer handles incoming concurrent Blocks
var bcServer chan []Block

func main() {
  err := godotenv.Load()
  if err != nil {
    log.Fatal(err)
  }

  bcServer = make(chan []Block)

  //Create genesis block

  t := time.Now()
  genesisBlock := Block{0, t.String(), 0, "", ""}
  spew.Dump(genesisBlock)
  Blockchain = append(Blockchain, genesisBlock)

  // Starting our TCP and serve TCP server. TCP server are similar to HTTP ones but whithout the browser component.
  server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
  if err != nil {
    log.Fatal(err)
  }
  defer server.Close()
  //This fires up the TCP server at por 9000.

  for {
    conn, err := server.Accept()

    if err != nil {
      log.Fatal(err)
    }

    go handleConn(conn)
  }
  //This is an infinite loop where we accept new connections. We will deal with each connection through a separate handler in a Go routine, so we can serve multiple connections concurently.


}

func handleConn(conn net.Conn) {
  defer conn.Close()

  io.WriteString(conn, "Enter a new BPM: ")

  scanner := bufio.NewScanner(conn)

  //take in BPM from stdin and add it to blockchain after conducting necessary validation
  go func() {
    for scanner.Scan() {
      bpm, err := strconv.Atoi(scanner.Text())

      if err != nil {
        log.Printf("%v not a number: %v", scanner.Text(), err)
        continue
      }

      newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], bpm)

      if err != nil {
        log.Println(err)
        continue
      }

      if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
        newBlockchain := append(Blockchain, newBlock)
        replaceChain(newBlockchain)
      }

      bcServer <- Blockchain
      io.WriteString(conn, "\nEnter a new BPM: ")
    }
  }()

  go func() {
    for {
      time.Sleep(30 * time.Second)
      output, err := json.Marshal(Blockchain)

      if err != nil {
        log.Fatal(err)
      }
      io.WriteString(conn, string(output))
    }
  }()

  for _ = range bcServer {
    spew.Dump(Blockchain)
  }
}
