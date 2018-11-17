package blockchain

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	//"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	//"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
	"encoding/json"
)

var WalletSuffix string

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int `json:"index"`
	Timestamp int64 `json:"timestamp"`
	Result       int `json:"result"`
	Validator	string 	`json:"validator"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prevhash"`
	Proof        uint64           `json:"proof"`
	Transactions []Transaction `json:"transactions"`
	CoinBase 	map[string]Account	`json:"coin_base"`
}

type Account struct {
	Addr 	 string
	Nonce    uint64
	Balance  uint64
}

type StateDB struct {
	Accounts 	map[string]Account
}

func NewState() *StateDB {
	return &StateDB{
		Accounts:   make(map[string]Account, 1),
	}
}

var State StateDB = StateDB{Accounts:   make(map[string]Account, 1)}

var tempState StateDB = StateDB{Accounts:   make(map[string]Account, 1)}

type Transaction struct {
	ID 		  string	`json:"id"`
	Sender    string 	`json:"sender"`
	Recipient string 	`json:"recipient"`
	Amount    uint64    `json:"amount"`
	AccountNonce uint64	`json:"account_nonce"`
	Data      []byte 	`json:"data"`
}

type TxPool struct {
	PengdingTx *TxsList //Unpackaged list of legitimate transactions
}



func NewTxPool() *TxPool {
	return &TxPool{
		PengdingTx:   NewTxsList(),
	}
}


//func (p *TxPool)Clear() bool {
//	if len(p.PengdingTx) == 0 {
//		return true
//	}
//	p.AllTx = make([]Transaction, 0)
//	return true
//}

// Blockchain is a series of validated Blocks
type Blockchain struct {
	Blocks []Block
	TxPool *TxPool
}

func (t *Blockchain) NewTransaction(sender string, recipient string, amount, nonce uint64, data []byte) *Transaction {
	transaction := new(Transaction)
	transaction.Sender = sender
	transaction.Recipient = recipient
	transaction.Amount = amount
	transaction.AccountNonce = nonce
	transaction.Data = data
	transaction.ID = calculateTransactionsHash(*transaction)

	return transaction
}

func PutStateDB(address string) int {
	if acc, ok := State.Accounts[address]; ok {
		if _, existed := tempState.Accounts[address]; !existed {
			tempState.Accounts[address] = acc
		}
	} else {
		return -1
	}

	return 0
}

func (t *Blockchain)AddTxPool(tx *Transaction) int {
	if ret := t.processTempTx(tx); ret == 0 {
		//t.TxPool.AllTx = append(t.TxPool.AllTx, *tx)
		t.TxPool.PengdingTx.Push(tx)
	} else {
		log.Println("invalid trx!")
	}

	return 0
}

func (t *Blockchain) LastBlock() Block {
	return t.Blocks[len(t.Blocks)-1]
}

func GetBalance(address string) uint64 {
	if acc, ok := State.Accounts[address]; ok {
		return acc.Balance
	}
	return 0
}

func GetTempDBNonce(address string) uint64 {
	if acc, ok := tempState.Accounts[address]; ok {
		return acc.Nonce
	}
	return 0
}

func GetNonce(address string) uint64 {
	if acc, ok := State.Accounts[address]; ok {
		return acc.Nonce
	}
	return 0
}

func AddBalance(address string, amount uint64) int{
	var ok bool
	acc, ok := State.Accounts[address]
	if ok {
		acc.Balance += 	amount
	} else {
		acc = Account{
			Addr:		address,
			Nonce:		0,
			Balance: 	amount,
		}
	}

	State.Accounts[address] = acc
	return 0
}

func SubBalance(address string, amount uint64) int{
	var ok bool
	acc, ok := State.Accounts[address]
	if ok {
		if acc.Balance > amount {
			acc.Nonce += 1
			acc.Balance -= 	amount
			State.Accounts[address] = acc
		} else {
			return -1
		}

	} else {
		return -2
	}

	return 0
}

func (t *Blockchain)processTempTx(tx *Transaction) int {
	if acc, ok := tempState.Accounts[tx.Sender]; ok {
		if acc.Balance < tx.Amount {
			return -1
		}

		if acc.Nonce >= tx.AccountNonce {
			return -2
		}
		acc.Nonce = tx.AccountNonce
		acc.Balance -= tx.Amount
		tempState.Accounts[tx.Sender] = acc
	} else {
		return -3
	}

	var ok bool
	acc, ok := State.Accounts[tx.Recipient]
	if ok {
		acc.Balance += 	tx.Amount
	} else {
		acc = Account{
			Addr:		tx.Recipient,
			Nonce:		0,
			Balance: 	tx.Amount,
		}
	}

	tempState.Accounts[tx.Recipient] = acc

	return 0
}

func (t *Blockchain)ProcessTx(newBlock Block) {
	for _, v := range newBlock.Transactions{
		if acc, ok := State.Accounts[v.Sender]; ok {
			acc.Nonce = v.AccountNonce
			acc.Balance -= v.Amount
			State.Accounts[v.Sender] = acc
		}

		AddBalance(v.Recipient, v.Amount)

		//for i, trx := range t.TxPool.AllTx {
		//	if trx.ID == v.ID {
		//		if i == len(t.TxPool.AllTx) {
		//			t.TxPool.AllTx =t.TxPool.AllTx[:i-1]
		//		} else {
		//			t.TxPool.AllTx = append(t.TxPool.AllTx[:i-1], t.TxPool.AllTx[i+1:]...)
		//		}
		//	}
		//}
		t.TxPool.PengdingTx.Delete(v.ID)

		log.Println("delete trx ", v.ID)

		tempState.Accounts[v.Sender] = State.Accounts[v.Sender]
	}

	//tempState = StateDB{Accounts:   make(map[string]Account, 0)}
	//t.TxPool.Clear()
}

func (t *Blockchain)PackageTx(newBlock *Block) {
	var txs []Transaction
	for _, v := range t.TxPool.PengdingTx.Txs {
		txs = append(txs, *v)
	}

	log.Printf("tx pool has %d trx\n", len(txs))

	(*newBlock).Transactions = txs
}

var BlockchainInstance Blockchain = Blockchain{
	TxPool : NewTxPool(),
}

var tempBlocks []Block

var NodeAccount string

var validators = make(map[string]int)

// candidateBlocks handles incoming blocks for validation
var candidateBlocks = make(chan Block)

// announcements broadcasts winning validator to all nodes
var announcements = make(chan string)

// brocast new Transaction
var BrocastNewTx = make(chan Transaction)

var mutex = &sync.Mutex{}


func Lock(){
	mutex.Lock()
}

func UnLock(){
	mutex.Unlock()
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -c chain -l %d -d %s -secio\" on a different terminal\n", listenPort+2, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -c chain -l %d -d %s\" on a different terminal\n", listenPort+2, fullAddr)
	}

	return basicHost, nil
}

func HandleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData(rw)
	go WriteData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func WaitingBlock() {
	for candidate := range candidateBlocks {
		mutex.Lock()
		tempBlocks = append(tempBlocks, candidate)
		mutex.Unlock()
	}
}

// pickWinner creates a lottery pool of validators and chooses the validator who gets to forge a block to the blockchain
// by random selecting from the pool, weighted by amount of tokens staked
func PickWinner() {
	for {
		if time.Now().Unix() % 10 == 0 {
			time.Sleep(time.Second)
			break
		}
	}

	//time.Sleep(15 * time.Second)	// long enough for peer node input result, it must add when peer node add
	mutex.Lock()
	temp := tempBlocks
	mutex.Unlock()

	lotteryPool := []string{}
	if len(temp) > 0 {

		// slightly modified traditional proof of stake algorithm
		// from all validators who submitted a block, weight them by the number of staked tokens
		// in traditional proof of stake, validators can participate without submitting a block to be forged
	OUTER:
		for _, block := range temp {
			// if already in lottery pool, skip
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			// lock list of validators to prevent data race
			mutex.Lock()
			setValidators := validators
			mutex.Unlock()

			k, ok := setValidators[block.Validator]
			if ok {
				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}
		}

		// randomly pick winner from lottery pool
		source := BlockchainInstance.LastBlock().Timestamp
		s := mrand.NewSource(source)
		r := mrand.New(s)
		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

		// add block of winner to blockchain and let all the other nodes know
		for _, block := range temp {
			if block.Validator == lotteryWinner {
				if block.Index > BlockchainInstance.LastBlock().Index {		// new block must higer than current last block
					mutex.Lock()
					if len(block.Transactions) > 0 {
						BlockchainInstance.ProcessTx(block)
					}
					BlockchainInstance.Blocks = append(BlockchainInstance.Blocks, block)
					mutex.Unlock()

					for _ = range validators {
						announcements <- "\nwinning validator: " + lotteryWinner + "\n"
					}
					log.Println("\nAfter pick, winning validator: " + lotteryWinner + "\n")

					bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
					if err != nil {

						log.Fatal(err)
					}
					// Green console color: 	\x1b[32m
					// Reset console color: 	\x1b[0m
					fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))

					break
				}
			}
		}
	}

	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}

// unused for the moment
func AnnounceWinner() {
	for {
		msg := <-announcements
		//mutex.Lock()
		//rw.WriteString(fmt.Sprintf("%s\n", msg))
		//rw.Flush()
		//mutex.Unlock()
		log.Println(msg)
	}
}

type stake struct {
	Index		string		`json:"index"`
	Balance		int			`json:"balance"`
}

// validator
func ReadData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}

		if strings.HasPrefix(str, "BrocastNewTx") {
			txStr := strings.TrimPrefix(str, "BrocastNewTx")
			tx := Transaction{}
			if err := json.Unmarshal([]byte(txStr), &tx); err != nil {
				log.Fatal(err)
			}

			log.Println(txStr)
			PutStateDB(tx.Sender)
			BlockchainInstance.AddTxPool(&tx)
		}


		// recv peer's stake info for pos account stake sync
		if strings.HasPrefix(str, "Stake") {
			stakeStr := strings.TrimPrefix(str, "Stake")
			stakeInfo := stake{}
			if err := json.Unmarshal([]byte(stakeStr), &stakeInfo); err != nil {
				log.Fatal(err)
			}

			log.Printf("%s stake %d token\n", stakeInfo.Index, stakeInfo.Balance)
			validators[stakeInfo.Index] = stakeInfo.Balance
		}

		// recv peer's new block for pos candidate block sync
		if strings.HasPrefix(str, "CandidateBlock") {
			BlockStr := strings.TrimPrefix(str, "CandidateBlock")

			var candidate Block
			if err := json.Unmarshal([]byte(BlockStr), &candidate); err != nil {
				log.Fatal(err)
			}
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}

		// recv newest blockchain
		if strings.HasPrefix(str, "Chain") {
			BlockChainStr := strings.TrimPrefix(str, "Chain")
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(BlockChainStr), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(BlockchainInstance.Blocks) {
				BlockchainInstance.Blocks = chain
				bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}

// producer
func WriteData(rw *bufio.ReadWriter) {

	// every 5 min, sync newest blockchain to peer
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockchainInstance.Blocks)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("Chain%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	go func() {
		for {
			msg := <-BrocastNewTx
			bytes, err := json.Marshal(msg)
			if err != nil {
				log.Println(err)
			}

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("BrocastNewTx%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
			log.Println(msg)
		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	fmt.Print("\nEnter token balance:")
	sendData, err := stdReader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	sendData = strings.Replace(sendData, "\n", "", -1)
	balance, err := strconv.Atoi(sendData)
	if err != nil {
		log.Fatal(err)
	}

	validators[NodeAccount] = balance

	stakeInfo := stake{
		Index:		NodeAccount,
		Balance:	balance,
	}

	bytes, err := json.Marshal(stakeInfo)
	if err != nil {
		log.Println(err)
	}

	mutex.Lock()
	rw.WriteString(fmt.Sprintf("Stake%s\n", string(bytes)))
	rw.Flush()
	mutex.Unlock()

	fmt.Println(validators)

	var ticker *time.Ticker
	for {
		if time.Now().Unix() % 10 == 0 {
			//time.Sleep(time.Second)
			ticker = time.NewTicker(10 * time.Second)
			break
		}
	}



	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Printf("ticked at %v\n", time.Now())
			}

			newBlock := GenerateBlock(BlockchainInstance.LastBlock(), 1, NodeAccount)

			//fmt.Print("\nEnter a new Result:")
			//sendData, err := stdReader.ReadString('\n')
			//if err != nil {
			//	log.Fatal(err)
			//}
			//
			//sendData = strings.Replace(sendData, "\n", "", -1)
			//_result, err := strconv.Atoi(sendData)
			//if err != nil {
			//	log.Fatal(err)
			//}
			//
			//newBlock := GenerateBlock(BlockchainInstance.LastBlock(), _result, NodeAccount)



			//if len(BlockchainInstance.TxPool.AllTx) > 0 {
				BlockchainInstance.PackageTx(&newBlock)
			//}

			if IsBlockValid(newBlock, BlockchainInstance.LastBlock()) {
				candidateBlocks <- newBlock
			}

			bytes, err := json.Marshal(newBlock)
			if err != nil {
				log.Println(err)
			}

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("CandidateBlock%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
		}
	}()


}



// make sure block is valid by checking index, and comparing the hash of the previous block
func IsBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if CalculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//calculateBlockHash returns the hash of Transaction information
func calculateTransactionsHash(trx Transaction) string {
	record := trx.Sender + trx.Recipient + strconv.FormatUint(trx.Amount, 10) + strconv.FormatUint(trx.AccountNonce, 10)
	return calculateHash(record)
}

//calculateBlockHash returns the hash of all block information
func CalculateBlockHash(block Block) string {
	record := strconv.Itoa(block.Index) + strconv.FormatInt(block.Timestamp,10) + strconv.Itoa(block.Result) + block.Validator + block.PrevHash
	return calculateHash(record)
}

// create a new block using previous block's hash
func GenerateBlock(oldBlock Block, result int, validator string) Block {
	var newBlock Block

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Now().Unix()
	newBlock.Result = result
	newBlock.Validator = validator
	newBlock.PrevHash = oldBlock.Hash
	newBlock.CoinBase = nil
	newBlock.Hash = CalculateBlockHash(newBlock)

	return newBlock
}
