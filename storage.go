package voting

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Election struct {
	Candidates  []Candidate `json:"candidates"`
	MyVote      []string    `json:"myvote" datastore:"-"`
	IsAdmin     bool        `json:"isadmin" datastore:"-"`
	Randomized  bool        `json:"randomized"`
	Limit       bool        `json:"limit"`
	Weight      int         `json:"weight"`
	Progress    bool        `json:"showProgress"`
	Description string      `json:"description"`
	Secondaries int         `json:"secondaries"`
}

type Candidate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Request     int    `json:"request"`
	Progress    int    `json:"progress" datastore:"-"`
	Section     string `json:"section"`
}

type Counter struct {
	Count int
}

type Vote struct {
	Vote     []string
	Election string
	Voter    string
}

var IntCodec = &memcache.Codec{
	Marshal: func(i interface{}) ([]byte, error) {
		return []byte(strconv.Itoa(*i.(*int))), nil
	},
	Unmarshal: func(b []byte, i interface{}) (err error) {
		y := i.(*int)
		*y, err = strconv.Atoi(string(b))
		return
	},
}

func MakeElectionKey(c appengine.Context, eid string) *datastore.Key {
	return datastore.NewKey(c, "Election", eid, 0, nil)
}

func MakeCounterKey(c appengine.Context, eid string, cand string) *datastore.Key {
	bound := fmt.Sprintf("c%d:%s", eid, cand)
	return datastore.NewKey(c, "Counter", bound, 0, nil)
}

func MakeVoteKey(c appengine.Context, eid string, voter string) *datastore.Key {
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

func ChangeCount(c appengine.Context, eid string, cand string, amt, lmt int) error {
	e := new(Counter)
	key := MakeCounterKey(c, eid, cand)
	return Mutate(c, key, e, func() (bool, error) {
		if lmt != 0 && e.Count >= lmt {
			return false, errors.New("Cannot exceed limit")
		}
		e.Count += amt
		memcache.IncrementExisting(c, key.Encode(), int64(amt))
		return true, nil
	})
}

func GetCount(c appengine.Context, eid string, cand string) (int, error) {
	key := MakeCounterKey(c, eid, cand)
	e := new(Counter)
	count := int(0)
	if _, err := IntCodec.Get(c, key.Encode(), &count); err == nil {
		return count, nil
	}
	err := Mutate(c, key, e, func() (bool, error) {
		return false, nil
	})
	IntCodec.Set(c, &memcache.Item{
		Key:    key.Encode(),
		Object: &e.Count,
	})
	return e.Count, err
}

func ChangeVote(c appengine.Context, eid, voter string, cands []string, limits []int) error {
	e := new(Vote)
	return Mutate(c, MakeVoteKey(c, eid, voter), e, func() (bool, error) {
		e.Voter = voter
		e.Election = eid
		for _, cand := range e.Vote {
			if err := ChangeCount(c, eid, cand, -1, 0); err != nil {
				return false, err
			}
		}
		e.Vote = cands
		for i, cand := range cands {
			if err := ChangeCount(c, eid, cand, +1, limits[i]); err != nil {
				return false, err
			}
		}
		return true, nil
	})
}

func GetVote(c appengine.Context, eid string, voter string) ([]string, error) {
	e := new(Vote)
	err := Mutate(c, MakeVoteKey(c, eid, voter), e, func() (bool, error) {
		return false, nil
	})
	return e.Vote, err
}

func GetVoters(c appengine.Context, eid string, elec *Election, w *csv.Writer) error {
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
		labels := make([]string, elec.Secondaries+1)
		for i := 0; i < 1+elec.Secondaries; i++ {
			if i < len(e.Vote) {
				labels[i] = lbls[e.Vote[i]]
			}
		}

		if err := w.Write([]string{e.Voter, strings.Join(labels, ";")}); err != nil {
			return err
		}
	}

	w.Flush()
	return nil
}
