package main

import (
	"context"
	"example.com/task149/tsnsched/internal/config"
	"example.com/task149/tsnsched/internal/httpapi"
	"example.com/task149/tsnsched/internal/metrics"
	"example.com/task149/tsnsched/internal/service"
	"example.com/task149/tsnsched/internal/store"
	"example.com/task149/tsnsched/internal/webui"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	c, smoke, err := config.Flags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	db, err := store.Open(c.Database)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	svc := service.New(db)
	svc.PeriodNS = c.NetworkPeriod
	svc.Guard.NetworkPeriodNS = c.NetworkPeriod
	svc.Guard.GuardBeforeNS = c.GuardBefore
	svc.Guard.GuardAfterNS = c.GuardAfter
	if err := svc.Recover(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if smoke {
		if err := runSmoke(svc); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("smoke-test: PASS")
		return
	}
	server := httpapi.New(svc, metrics.New())
	server.Page = webui.Load("web/index.html")
	fmt.Printf("tsnsched listening on %s\n", c.Address)
	if err := http.ListenAndServe(c.Address, server.Handler()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runSmoke(svc *service.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, n := range []struct{ id, name string }{{"n1", "edge"}, {"n2", "switch-a"}, {"n3", "switch-b"}, {"n4", "sink"}} {
		if err := svc.AddNode(ctx, modelNode(n.id, n.name)); err != nil {
			return err
		}
	}
	for _, p := range []struct{ id, node string }{{"p1", "n1"}, {"p2", "n2"}, {"p3", "n3"}, {"p4", "n4"}} {
		if err := svc.AddPort(ctx, modelPort(p.id, p.node)); err != nil {
			return err
		}
	}
	for _, l := range []struct{ id, a, b string }{{"l1", "p1", "p2"}, {"l2", "p2", "p3"}, {"l3", "p3", "p4"}} {
		if err := svc.AddLink(ctx, modelLink(l.id, l.a, l.b)); err != nil {
			return err
		}
	}
	if err := svc.AddStream(ctx, modelStream("s1", []string{"p1", "p2", "p3", "p4"})); err != nil {
		return err
	}
	s, err := svc.CreateDraft(ctx, "smoke", 1)
	if err != nil {
		return err
	}
	if !s.Validation.Valid {
		return fmt.Errorf("smoke draft invalid: %v", s.Validation.Violations)
	}
	if err = svc.Commit(ctx, "smoke", s.Draft.ID); err != nil {
		return err
	}
	_, err = svc.Active(ctx, "smoke")
	return err
}
