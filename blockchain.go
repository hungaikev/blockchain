package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

//Block contains data that will be written to the blockchain.
type Block struct {
	Pos       int
	Data      BookCheckout
	Timestamp string
	Hash      string
	PrevHash  string
}

func (b *Block) generateHash() {
	// get string val of the Data
	bytes, _ := json.Marshal(b.Data)
	// concate the dataset
	data := string(b.Pos) + b.Timestamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

// CreateBlock to create a new block
func CreateBlock(prevBlock *Block, checkoutItem BookCheckout) *Block {
	block := &Block{}
	block.Pos = prevBlock.Pos + 1
	block.Timestamp = time.Now().String()
	block.Data = checkoutItem
	block.PrevHash = prevBlock.Hash
	block.generateHash()

	return block
}

// GenesisBlock is a function to create the first Block in the chain
func GenesisBlock() *Block {
	return CreateBlock(&Block{}, BookCheckout{IsGenesis: true})
}

// NewBlockchain function returns a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

// validBlock function
func validBlock(block, prevBlock *Block) bool {
	// Confirm the hashes
	if prevBlock.Hash != block.PrevHash {
		return false
	}

	// Confirm the blocks has is valid
	if !block.validateHash(block.Hash) {
		return false
	}

	// Check the position to confirm its been incremented.
	if prevBlock.Pos+1 != block.Pos {
		return false
	}
	return true
}

// validateHash function
func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

//BookCheckout contains data for a checked out book
type BookCheckout struct {
	BookID       string `json:"book_id"`
	User         string `json:"user"`
	CheckoutDate string `json:"checkout_date"`
	IsGenesis    bool   `json:"is_genesis"`
}

//Book contains data for a sample book
type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	PublishDate string `json:"publish_date"`
	ISBN        string `json:"isbn"`
}

//Blockchain is an ordered list of blocks
type Blockchain struct {
	blocks []*Block
}

//BlockChain is a global variable that'll return the mutated Blockchain struct
var BlockChain *Blockchain

//AddBlock adds a Block to a Blockchain
func (bc *Blockchain) AddBlock(data BookCheckout) {
	// get previosu block
	prevBlock := bc.blocks[len(bc.blocks)-1]
	// create new block
	block := CreateBlock(prevBlock, data)
	//validate integrity of blocks
	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	}
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", " ")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	// write JSON string
	io.WriteString(w, string(jbytes))
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutItem BookCheckout
	if err := json.NewDecoder(r.Body).Decode(&checkoutItem); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write Block: %v", err)
		w.Write([]byte("could not write block"))
		return
	}

	// create block
	BlockChain.AddBlock(checkoutItem)
	resp, err := json.MarshalIndent(checkoutItem, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not write block"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func newBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("would not create: %v", err)
		w.Write([]byte("could not create new Book"))
		return
	}
	// We'll create an ID, concatenating the ISDBand publish date
	// This isn't an efficient way but it serves for this tutorial
	resp, err := json.MarshalIndent(book, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not save book data"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func main() {

	// initialize the blockchain and store in var
	BlockChain = NewBlockchain()

	// register router
	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBook).Methods("POST")

	// dump the state of the Blockchain to the console
	go func() {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
	}()

	log.Println("Listening on port 3000")

	log.Fatal(http.ListenAndServe(":3000", r))

}
