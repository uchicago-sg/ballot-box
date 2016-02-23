package voting

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
)

type VoteSubmission struct {
	Candidates []string `json:"candidates"`
}

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	eid, verb, ext := RouteURL(r.URL.Path)

	k := MakeElectionKey(c, eid)
	e := new(Election)

	u := user.Current(c)

	if u == nil || strings.Split(u.Email, "@")[1] != "uchicago.edu" {
		url, err := user.LoginURL(c, r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		url += "&hd=uchicago.edu"
		http.Redirect(w, r, url, 303)
		return
	}

	if u.Admin && eid == "" {

		w.Header().Add("Content-type", "text/html")
		fmt.Fprintf(w,
			"<!doctype html><html>"+
				"<head>"+
				"<title>Student Government</title>"+
				"<meta name='viewport' content='width=device-width, initial-scale=1'>"+
				"<link rel='stylesheet' href='https://maxcdn.bootstrapcdn.com/"+
				"bootstrap/3.3.4/css/bootstrap.min.css'>"+
				"<link rel='stylesheet' href='/styles.css'>"+
				"</head>"+
				"<body class='container'><form method='post' action='/%d?create=y'>"+
				"<input type='submit' value='create'/></form></body>"+
				"</html>", rand.Int31(),
		)
		return

	} else if r.Method == "POST" && verb == "" && u.Admin {

		err := json.NewDecoder(r.Body).Decode(&e)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if e.Secondaries < 0 {
			e.Secondaries = 0
		}
		if e.Secondaries >= len(e.Candidates) {
			e.Secondaries = len(e.Candidates) - 1
		}

		if _, err := datastore.Put(c, k, e); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	} else if r.Method == "POST" && verb == "vote" {

		if err := datastore.Get(c, k, e); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		subs := new(VoteSubmission)

		err := json.NewDecoder(r.Body).Decode(&subs)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		cands := subs.Candidates
		for len(cands) <= e.Secondaries {
			cands = append(cands, "")
		}
		cands = cands[:e.Secondaries+1]
		limits := make([]int, len(cands))

		for _, c := range e.Candidates {
			for i, cand := range cands {
				if c.ID == cand {
					if e.Weight == 0 {
						limits[i] = c.Request
					} else {
						limits[i] = c.Request / e.Weight
					}
					break
				}
			}
		}

		err = datastore.RunInTransaction(c, func(c appengine.Context) error {
			if err := ChangeVote(c, eid, u.Email, cands, limits); err != nil {
				return err
			}

			return nil
		}, &datastore.TransactionOptions{XG: true})

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	} else if r.Method == "GET" && verb == "results" && ext == "csv" && u.Admin {

		if err := datastore.Get(c, k, e); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Add("Content-Disposition",
			"attachment; filename=\"results.csv\"")
		w.Header().Add("Content-type", "text/csv")

		if err := GetVoters(c, eid, e, csv.NewWriter(w)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		return

	} else if r.Method == "GET" && verb == "" && eid != "" {

		if err := datastore.Get(c, k, e); err != nil {
			if err == datastore.ErrNoSuchEntity {
				if u.Admin {
					if _, err := datastore.Put(c, k, e); err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
				} else {
					http.NotFound(w, r)
					return
				}
			} else {
				http.Error(w, err.Error(), 500)
				return
			}
		}

	} else {

		http.NotFound(w, r)
		return

	}

	for i, candidate := range e.Candidates {
		if candidate.Request > 0 {
			prog, err := GetCount(c, eid, candidate.ID)
			e.Candidates[i].Progress = prog
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		} else {
			e.Candidates[i].Progress = 0
		}
	}

	vote, err := GetVote(c, eid, u.Email)
	e.MyVote = vote
	for len(e.MyVote) <= e.Secondaries {
		e.MyVote = append(e.MyVote, "")
	}
	e.IsAdmin = u.Admin
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	j, err := json.Marshal(e)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if ext == "json" {
		fmt.Fprintf(w, "%s", j)
	} else {
		w.Header().Add("Content-type", "text/html")
		fmt.Fprintf(w, "%s",
			"<!doctype html><html ng-app='voteApp'>"+
				"<head>"+
				"<title>Student Government</title>"+
				"<meta name='viewport' content='width=device-width, initial-scale=1'>"+
				"<link rel='stylesheet' href='https://maxcdn.bootstrapcdn.com/"+
				"bootstrap/3.3.4/css/bootstrap.min.css'>"+
				"<link rel='stylesheet' href='/styles.css'>"+
				"<script src='https://cdnjs.cloudflare.com/ajax/libs/"+
				"angular.js/1.3.15/angular.min.js'></script>"+
				"<script src='https://cdnjs.cloudflare.com/ajax/libs/angular-filter/"+
				"0.5.8/angular-filter.min.js'></script>"+
				"<script src='https://cdnjs.cloudflare.com/ajax/libs/"+
				"angular-ui-bootstrap/0.13.0/ui-bootstrap.min.js'></script>"+
				"<script src='https://cdnjs.cloudflare.com/ajax/libs/showdown/1.0.1/"+
				"showdown.min.js'></script>"+
				"<script src='/main.js'></script>"+
				"<script type='text/javascript'>var _DATA="+string(j)+"</script>"+
				"</head>"+
				"<body ng-include='\"/root.tpl\"'></body>"+
				"</html>",
		)
	}

}
