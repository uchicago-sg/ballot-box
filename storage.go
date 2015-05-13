package voting

import (
	"appengine"
	"appengine/datastore"
	"encoding/csv"
	"errors"
	"fmt"
)

type Election struct {
	Candidates []Candidate `json:"candidates"`
	MyVote     string      `json:"myvote" datastore:"-"`
	Randomized bool        `json:"randomized"`
	Limit      bool        `json:"limit"`
	Weight     int         `json:"weight"`
}

type Candidate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Request     int    `json:"request"`
	Progress    int    `json:"progress" datastore:"-"`
}

type Counter struct {
	Count int
}

type Vote struct {
	Vote     string
	Election int64
	Voter    string
}

func MakeElectionKey(c appengine.Context, eid int64) *datastore.Key {
	return datastore.NewKey(c, "Election", "", eid, nil)
}

func MakeCounterKey(c appengine.Context, eid int64, cand string) *datastore.Key {
	bound := fmt.Sprintf("c%d:%s", eid, cand)
	return datastore.NewKey(c, "Counter", bound, 0, nil)
}

func MakeVoteKey(c appengine.Context, eid int64, voter string) *datastore.Key {
	bound := fmt.Sprintf("v%d:%s", eid, voter)
	return datastore.NewKey(c, "Vote", bound, 0, nil)
}

func Mutate(c appengine.Context, key *datastore.Key, ent interface{},
	mut func() (bool, error)) error {

	err := datastore.Get(c, key, ent)
	if err != nil && err != datastore.ErrNoSuchEntity {
		return err
	}

	save, err := mut()
	if err != nil {
		return err
	}

	if save {
		if _, err := datastore.Put(c, key, ent); err != nil {
			return err
		}
	}

	return nil
}

func ChangeCount(c appengine.Context, eid int64, cand string, amt, lmt int) error {
	e := new(Counter)
	return Mutate(c, MakeCounterKey(c, eid, cand), e, func() (bool, error) {
		if lmt != 0 && e.Count >= lmt {
			return false, errors.New("Cannot exceed limit")
		}
		e.Count += amt
		return true, nil
	})
}

func GetCount(c appengine.Context, eid int64, cand string) (int, error) {
	e := new(Counter)
	err := Mutate(c, MakeCounterKey(c, eid, cand), e, func() (bool, error) {
		return false, nil
	})
	return e.Count, err
}

func ChangeVote(c appengine.Context, eid int64, voter, cand string, limit int) error {
	e := new(Vote)
	return Mutate(c, MakeVoteKey(c, eid, voter), e, func() (bool, error) {
		if e.Vote == cand {
			return false, nil
		}
		e.Voter = voter
		e.Election = eid
		if e.Vote != "" {
			if err := ChangeCount(c, eid, e.Vote, -1, 0); err != nil {
				return false, err
			}
		}
		e.Vote = cand
		if err := ChangeCount(c, eid, cand, +1, limit); err != nil {
			return false, err
		}
		return true, nil
	})
}

func GetVote(c appengine.Context, eid int64, voter string) (string, error) {
	e := new(Vote)
	err := Mutate(c, MakeVoteKey(c, eid, voter), e, func() (bool, error) {
		return false, nil
	})
	return e.Vote, err
}

func GetVoters(c appengine.Context, eid int64, elec *Election, w *csv.Writer) error {
	q := datastore.NewQuery("Vote").Filter("Election =", eid).Order("Vote")
	lbls := make(map[string]string)
	for _, c := range elec.Candidates {
		lbls[c.ID] = c.Name
	}

	if err := w.Write([]string{"email", "selection"}); err != nil {
		return err
	}

	for t := q.Run(c); ; {
		var e Vote
		_, err := t.Next(&e)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := w.Write([]string{e.Voter, lbls[e.Vote]}); err != nil {
			return err
		}
	}

	w.Flush()
	return nil
}
