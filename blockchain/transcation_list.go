package blockchain

/**
** 交易列表结构体
 */import (
	"sync"
	"fmt"
)

type TxsList struct {
	Txs map[string]*Transaction

	mux sync.RWMutex
}

func NewTxsList() *TxsList {
	return &TxsList{Txs: make(map[string]*Transaction, 0)}
}

//删除一笔交易
func (t *TxsList) Delete(hash string) {
	t.mux.Lock()
	defer t.mux.Unlock()

	delete(t.Txs, hash)
}

//检查是否有重复的数据
func (t *TxsList) Same(hash string) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if value := t.Txs[hash]; nil != value {
		return true
	}

	return false
}

/**
** 入栈操作
 */
func (t *TxsList) Push(tx *Transaction) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if _, ok := t.Txs[tx.ID]; ok {
		return
	}
	t.Txs[tx.ID] = tx
}

/**
** 出栈操作，取出后立即从列表移除
 */
func (t *TxsList) Pull() (tx *Transaction) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if len(t.Txs) == 0 {
		return nil
	}
	var tt *Transaction
	for k, v := range t.Txs {
		tt = v
		delete(t.Txs, k)
		return tt
	}
	return nil
}

func (t *TxsList) Copy(txs *TxsList) {
	txs.mux.RLock()
	defer txs.mux.RUnlock()
	for k, v := range txs.Txs {
		t.Txs[k] = v
	}
}

func (t *TxsList) Show() {
	t.mux.RLock()
	defer t.mux.RUnlock()
	for _, v := range t.Txs {
		fmt.Println("ID:", v.ID)
		fmt.Println("Sender:", v.Sender)
		fmt.Println("Recipient:", v.Recipient)
		fmt.Println("Amount:", v.Amount)
		fmt.Println("Nonce:", v.AccountNonce)
		fmt.Println("Data:", v.Data)
	}
}