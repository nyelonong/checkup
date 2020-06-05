package checkup

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type (
	Dependency struct {
		API      []API `yaml:"api"`
		Database struct {
			Postgres []Postgres `yaml:"postgres"`
			Redis    []Rediss   `yaml:"redis"`
		} `yaml:"database"`
		Grpc []GRPC `yaml:"grpc"`
	}

	API struct {
		Endpoint   string `yaml:"endpoint"`
		StatusCode int    `yaml:"statuscode"`
		Timeout    int64  `yaml:"timeout"`
	}

	GRPC struct {
		Host    string `yaml:"host"`
		Timeout int64  `yaml:"timeout"`
	}

	Postgres struct {
		Conn string `yaml:"conn"`
	}

	Rediss struct {
		Conn string `yaml:"conn"`
	}

	Module struct {
	 	Dep Dependency
	}
)

var (
	isDebug bool
)

func readDependencies(file string) (Dependency, error) {
	var dest Dependency
	data, err := ioutil.ReadFile(file)
	if err != nil {
		debugLog(err)
		return dest, err
	}

	if err := yaml.Unmarshal(data, &dest); err != nil {
		debugLog(err)
		return dest, err
	}

	return dest, nil
}

func debugLog(err error) {
	if isDebug {
		log.Println("[CHECKUP][DEBUG]", err)
	}
}

func infoLog(info string) {
	log.Println("[CHECKUP][INFO]", info)
}

func New(file string, debug bool) (*Module, error) {
	isDebug = debug

	dep, err := readDependencies(file)
	if err != nil {
		debugLog(err)
		return nil, ErrReadFile
	}

	log.Println("[CHECKUP][INFO] Dependencies file successfully loaded from", file)
	return &Module{
		Dep: dep,
	}, nil
}

func (m Module) Checkup() error {
	var group = sync.WaitGroup{}
	var errs []error

	if len(m.Dep.API) > 0 {
		group.Add(1)
		go func() {
			defer func() {
				group.Done()
			}()

			if err := m.checkupAPI(); err != nil {
				errs = append(errs, err)
			}
		}()
	}

	if len(m.Dep.Database.Postgres) > 0 {
		group.Add(1)
		go func() {
			defer func() {
				group.Done()
			}()

			if err := m.checkupPostgres(); err != nil {
				errs = append(errs, err)
			}
		}()
	}

	if len(m.Dep.Database.Redis) > 0 {
		group.Add(1)
		go func() {
			defer func() {
				group.Done()
			}()

			if err := m.checkupRedis(); err != nil {
				errs = append(errs, err)
			}
		}()
	}

	if len(m.Dep.Grpc) > 0 {
		group.Add(1)
		go func() {
			defer func() {
				group.Done()
			}()

			if err := m.checkupGrpc(); err != nil {
				errs = append(errs, err)
			}
		}()
	}

	group.Wait()
	if len(errs) > 0 {
		for _, err := range errs {
			debugLog(err)
		}

		return ErrCheckupFailed
	}

	infoLog("All dependencies is healthy. Ready to go.")
	return nil
}

func (m Module) checkupAPI() error {
	for _, api := range m.Dep.API {
		sc, err := doHTTPCall(api)
		if err != nil {
			return ErrHTTPCall
		}

		if sc != api.StatusCode {
			return ErrInvalidStatusCode
		}
		infoLog(fmt.Sprintf("[API] %s is health", api.Endpoint))
	}

	return nil
}

func (m Module) checkupPostgres() error {
	for _, psql := range m.Dep.Database.Postgres {
		db, err := sqlx.Connect("postgres", psql.Conn)
		if err != nil {
			debugLog(err)
			return ErrFailedConnectPostgres
		}

		if err := db.Ping(); err != nil {
			debugLog(err)
			return ErrFailedConnectPostgres
		}

		infoLog(fmt.Sprintf("[Postgres] %s is health", psql.Conn))
		db.Close()
	}

	return nil
}

func (m Module) checkupRedis() error {
	for _, red := range m.Dep.Database.Redis {
		r, err := redis.Dial("tcp", red.Conn)
		if err != nil {
			debugLog(err)
			return ErrFailedConnectRedis
		}

		if _, err := r.Do("PING"); err != nil {
			debugLog(err)
			return ErrFailedConnectRedis
		}

		infoLog(fmt.Sprintf("[Redis] %s is health", red.Conn))
	}

	return nil
}

func (m Module) checkupGrpc() error {
	for _, g := range m.Dep.Grpc {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(g.Timeout))
		defer cancel()

		conn, err := grpc.DialContext(ctx, g.Host, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			debugLog(err)
			return ErrFailedConnectGrpc
		}
		defer conn.Close()

		infoLog(fmt.Sprintf("[gRPC] %s is health", g.Host))
	}

	return nil
}
