package kas

import (
	"flag"
	"math/rand/v2"
	"slices"
	"testing"

	"pgregory.net/rapid"
)

/*
This model could be improved by having the signTxKas sign a batch of withdrawals at once, since this is closer to the spec and how
the real implementation will look. It will require a refactor to the outpoint representation. This is detailed below.
*/

// go test -count=1 -v ./...
func TestModel(t *testing.T) {
	// this runs for approx 2 mins
	_ = flag.Set("rapid.checks", "2000000")
	_ = flag.Set("rapid.steps", "200")
	rapid.Check(t, func(t *rapid.T) {
		rSeed := rapid.Uint64().Draw(t, "rSeed")
		m := initModel(t, rSeed)
		t.Repeat(map[string]func(*rapid.T){
			"": func(t *rapid.T) {
				// check invariants etc
				diff := m.qEscrow() - m.qMint()
				if diff < 0 {
					t.Fatalf("escrow < mint: %d < %d", m.qEscrow(), m.qMint())
				}
				if m.qHubWithdrawalsIncomplete() < diff {
					t.Fatalf("hub in flight withdrawals + mint < escrow: %d + %d < %d", m.qHubWithdrawalsIncomplete(), m.qMint(), m.qEscrow())
				}
			},
			"deposit": func(t *rapid.T) {
				x := rapid.IntRange(1, 100).Draw(t, "deposit x")
				m.deposit(x)
			},
			"withdraw": func(t *rapid.T) {
				x := rapid.IntRange(1, 100).Draw(t, "withdraw x")
				m.withdraw(x)
			},
			"signTxKas": func(t *rapid.T) {
				if 0 < len(m.hubWithdrawals) {
					maxL := len(m.hubWithdrawals) - 1
					newL := rapid.IntRange(0, maxL).Draw(t, "newL")
					m.signTxKas(newL)
				}
			},
			"deliverTxKas": func(t *rapid.T) {
				candidates := m.qUndeliveredTxKas()
				if 0 < len(candidates) {
					ix := rapid.SampledFrom(candidates).Draw(t, "deliverTxKas txKasIx")
					m.deliverTxKas(ix)
				}
			},
			"signTxHub": func(t *rapid.T) {
				maxI := len(m.txsKas) - 1
				if 0 <= maxI {
					i := rapid.IntRange(0, maxI).Draw(t, "signTxHub txKasIx")
					oldO := rapid.IntRange(0, len(m.kasUTXO)-1).Draw(t, "signTxHub oldO")
					m.signTxHub(i, oldO)
				}
			},
			"deliverTxHub": func(t *rapid.T) {
				candidates := m.qUndeliveredTxHub()
				if 0 < len(candidates) {
					i := rapid.SampledFrom(candidates).Draw(t, "deliverTxHub txHubIx")
					m.deliverTxHub(i)
				}
			},
		})
	})
}

type Model struct {
	///////////////////////
	// model bookkeeping

	t               *rapid.T
	r               *rand.Rand
	txsKasDelivered []bool // just a model optimization to avoid re checking txs again and again (not semantically important)
	txsHubDelivered []bool // just a model optimization to avoid re checking txs again and again (not semantically important)

	///////////////////////
	// algorithm data

	kasUTXO        []int  // deposit amounts set
	kasUTXOSpent   []bool // was it spent?
	hubMint        int    // minted amount
	hubWithdrawals []int  // withdrawal amounts queue
	O              int
	L              int
	txsKas         []TxKas // signed txs
	txsKasAccepted []bool  // if accepted, it can be confirmed
	txsHub         []TxHub // signed txs
}

const seededKas = 1

func initModel(t *rapid.T, seed uint64) *Model {
	return &Model{
		t:               t,
		r:               rand.New(rand.NewPCG(seed, 0)),
		txsKasDelivered: []bool{},
		txsHubDelivered: []bool{},

		kasUTXO:        []int{seededKas},
		kasUTXOSpent:   []bool{false},
		hubMint:        1,
		hubWithdrawals: []int{},
		O:              0, // init to zero index corresponding to seed UTXO
		L:              -1,
		txsKas:         []TxKas{},
		txsKasAccepted: []bool{},
		txsHub:         []TxHub{},
	}
}

////////////////////////////
// Queries

func (m *Model) qEscrow() int {
	sum := 0
	for i, utxo := range m.kasUTXO {
		if !m.kasUTXOSpent[i] {
			sum += utxo
		}
	}
	return sum
}

func (m *Model) qMint() int {
	return m.hubMint
}

func (m *Model) qHubWithdrawalsIncomplete() int {
	sum := 0
	for _, w := range m.hubWithdrawals[m.L+1:] {
		sum += w
	}
	return sum
}

func (m *Model) qRandomUnspentUTXOs() []int {
	ixs := make([]int, len(m.kasUTXO))
	for i := range ixs {
		ixs[i] = i
	}
	m.r.Shuffle(len(ixs), func(i, j int) {
		ixs[i], ixs[j] = ixs[j], ixs[i]
	})

	n := m.r.IntN(len(ixs) + 1)
	var ret []int
	for _, i := range ixs {
		if !m.kasUTXOSpent[i] {
			ret = append(ret, i)
		}
		if len(ret) == n {
			break
		}
	}
	return ret
}

func (m *Model) qUndeliveredTxKas() []int {
	var ret []int
	for i, d := range m.txsKasDelivered {
		if !d {
			ret = append(ret, i)
		}
	}
	return ret
}

func (m *Model) qUndeliveredTxHub() []int {
	var ret []int
	for i, d := range m.txsHubDelivered {
		if !d {
			ret = append(ret, i)
		}
	}
	return ret
}

////////////////////////////
// Actions

// spend UTXO on kas to unescrow tokens
type TxKas struct {
	utxos    []int // indexes to utxos
	amount   int   // how much withdrawing user gets
	newL     int   // completed withdrawal queue index
	outpoint int   // new anchor (populated on acceptance)
}

// update hub with new bookkeeping
type TxHub struct {
	confirmedTxKas int
	oldO           int
	newO           int
	newL           int
}

// user
func (m *Model) deposit(x int) {
	m.kasUTXO = append(m.kasUTXO, x)
	m.kasUTXOSpent = append(m.kasUTXOSpent, false)
	// for now we assume a trivial relay to the hub
	m.hubMint += x
}

// user
func (m *Model) withdraw(x int) {
	m.hubWithdrawals = append(m.hubWithdrawals, x)
	m.hubMint -= x
}

// validator :: unescrow funds to finish withdrawal
func (m *Model) signTxKas(newL int) {
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// relayer section (modeled here for convenience)

	txKas := TxKas{
		utxos:  m.qRandomUnspentUTXOs(),
		amount: m.hubWithdrawals[newL],
		newL:   newL,
		// TODO: model limitation: we currently model outpoint as being defined on the TX acceptance, however in Kaspa
		// 	they are defined on the TX creation to enable client TX linking. To model that we would need to change the repr here.
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// validator section. Verify!

	if !slices.Contains(txKas.utxos, m.O) {
		m.t.Logf("proposed TX does not spend O: %d", m.O)
		return
	}

	expectedValue := 0
	for _, w := range m.hubWithdrawals[m.L+1 : txKas.newL+1] {
		expectedValue += w
	}
	if expectedValue != txKas.amount {
		m.t.Logf("proposed TX does not credit the right amount to the user(s): expect %d != tx credit %d", expectedValue, txKas.amount)
		return
	}

	proposedSpend := 0
	for _, utxo := range txKas.utxos {
		proposedSpend += m.kasUTXO[utxo]
	}
	if proposedSpend < txKas.amount+seededKas {
		m.t.Logf("proposed TX does not spend the right amount needed to send change back to escrow (which is needed to have a new anchor): proposed %d < required %d", proposedSpend, txKas.amount+seededKas)
		return
	}

	m.txsKas = append(m.txsKas, txKas)
	m.txsKasAccepted = append(m.txsKasAccepted, false)
	m.txsKasDelivered = append(m.txsKasDelivered, false)
}

// relayer
func (m *Model) deliverTxKas(i int) {
	if m.txsKasDelivered[i] {
		panic("model bug: tx kas already delivered")
	}
	tx := m.txsKas[i]
	m.txsKasDelivered[i] = true
	for i := range tx.utxos {
		if m.kasUTXOSpent[i] {
			m.t.Logf("Rejected TX, UTXO already spent: utxo: %d", i)
			return
		}
	}
	sum := 0
	for _, utxo := range tx.utxos {
		sum += m.kasUTXO[utxo]
		m.kasUTXOSpent[utxo] = true
	}
	remainder := sum - tx.amount
	if remainder < seededKas {
		// it should be impossible as there is always 1 kaspa seeded
		panic("spec bug: not enough escrow")
	}
	m.txsKasAccepted[i] = true
	m.kasUTXO = append(m.kasUTXO, remainder)
	m.kasUTXOSpent = append(m.kasUTXOSpent, false)
	m.txsKas[i].outpoint = len(m.kasUTXO) - 1
}

// validator :: update hub to track confirmations
func (m *Model) signTxHub(txKasIx int, oldO int) {
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// relayer section (modeled here for convenience)
	txKas := m.txsKas[txKasIx]
	txHub := TxHub{
		confirmedTxKas: txKasIx,
		oldO:           oldO,
		newO:           txKas.outpoint,
		newL:           txKas.newL,
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// validator section. Verify!

	if txHub.confirmedTxKas < 0 || len(m.txsKas) <= txHub.confirmedTxKas {
		// kas tx not found
		return
	}

	txKas = m.txsKas[txHub.confirmedTxKas]
	if !m.txsKasAccepted[txHub.confirmedTxKas] {
		m.t.Logf("Reject signing: Kasp TX not confirmed: kas tx: %d", txHub.confirmedTxKas)
		return
	}
	if txHub.oldO != m.O {
		m.t.Logf("Reject signing: given O not the current anchor: given %d != hub %d", txHub.oldO, m.O)
		return
	}
	if txHub.newO != txKas.outpoint {
		m.t.Logf("Reject signing: confirmed TX did not have given outpoint: given %d != confirmed tx %d", txHub.newO, txKas.outpoint)
		return
	}
	if txHub.newL != txKas.newL {
		m.t.Logf("Reject signing: confirmed TX did not handle withdrawals up to given newL: given %d != confirmed tx %d", txHub.newL, txKas.newL)
		return
	}
	if !slices.Contains(txKas.utxos, txHub.oldO) {
		m.t.Logf("Reject signing: confirmed TX did not spend O: hub %d != confirmed tx %d", txHub.oldO, txKas.utxos)
		return
	}

	m.txsHub = append(m.txsHub, txHub)
	m.txsHubDelivered = append(m.txsHubDelivered, false)
}

// relayer
func (m *Model) deliverTxHub(i int) {
	if m.txsHubDelivered[i] {
		panic("model bug: tx hub already delivered")
	}
	tx := m.txsHub[i]
	m.txsHubDelivered[i] = true
	if tx.oldO == m.O { // cas
		m.O = tx.newO
		m.L = tx.newL
	}
}
