package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	// "time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/olivere/jobqueue"
	"github.com/olivere/jobqueue/mysql"
)

func getUser(c echo.Context) error {
	// User ID from path `users/:id
	id := c.Param("id")
	pv := c.ParamValues()
	pn := c.ParamNames()

	fv := c.FormValue("address")

	name := c.QueryParam("name")
	params := c.QueryParams()
	queries := c.QueryString()

	return c.JSON(http.StatusOK, []interface{}{id, pv, pn, fv, name, params, queries, c.Request().Header})
}

type (
	User struct {
		Path  string                 `param:"path"`
		Name  map[string]interface{} `json:"name" validate:"dive"`
		Email string                 `json:"email" xml:"email" form:"email" query:"email"`
	}

	Header struct {
		Type string `header:"content-type"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func postUser(c echo.Context) error {
	u := new(User)
	h := new(Header)

	err := c.Bind(u)

	binder := &echo.DefaultBinder{}

	hErr := binder.BindHeaders(c, h)

	if err = c.Validate(u); err != nil {
		return err
	}

	if err != nil || hErr != nil {
		return err
	}

	header := c.Request().Header
	return c.JSONPretty(http.StatusCreated, []interface{}{u, h, header}, " ")
	// or
	// return c.XML(http.StatusCreated, u)
}

func queuejob() {
	// Create a MySQL-based persistent backend.
	store, err := mysql.NewStore("root@(127.0.0.1:3306)/skyhotel?loc=UTC&parseTime=true")
	if err != nil {
		panic(err)
	}

	// Create a manager with the MySQL store and 10 concurrent workers.
	m := jobqueue.New(
		jobqueue.SetStore(store),
		// jobqueue.SetConcurrency(0, 1),
	)

	fmt.Println(m)

	// Register one or more topics and their processor
	m.Register("Room Type", func(job *jobqueue.Job) error {
		time.Sleep(15 * time.Second)
		fmt.Println("Method1", job)
		return nil
	})

	
	// Register one or more topics and their processor
	m.Register("Plan", func(job *jobqueue.Job) error {
		time.Sleep(15 * time.Second)
		fmt.Println("Method1", job)
		return nil
	})

	// Start the manager
	// err = m.Start()
	// if err != nil {
	// 	panic(err)
	// }

	// Add a job: It'll be added to the store and processed eventually.
	job := &jobqueue.Job{Topic: "Room Type", Rank: 1, MaxRetry: 3, CorrelationGroup: "1", Args: []interface{}{640, 480}}
	err = m.Add(context.Background(), job)
	if err != nil {
		fmt.Println("Add failed")
		return
	}
	fmt.Println("Job added")

	job2 := &jobqueue.Job{Topic: "Plan", Rank: 2, MaxRetry: 3, CorrelationGroup: "2", Args: []interface{}{111, 222}}
	err = m.Add(context.Background(), job2)
	if err != nil {
		fmt.Println("Add failed2")
		return
	}
	fmt.Println("Job added2")

	startWorker()
}

func startWorker() {
	// Create a MySQL-based persistent backend.
	store, err := mysql.NewStore("root@(127.0.0.1:3306)/skyhotel?loc=UTC&parseTime=true")
	if err != nil {
		panic(err)
	}

	// Create a manager with the MySQL store and 10 concurrent workers.
	m := jobqueue.New(
		jobqueue.SetStore(store),
		jobqueue.SetConcurrency(0, 1),
		jobqueue.SetConcurrency(1, 1),
		jobqueue.SetConcurrency(2, 2),
	)

	// Register one or more topics and their processor
	m.Register("Room Type", func(job *jobqueue.Job) error {
		time.Sleep(15 * time.Second)
		fmt.Println("Method1", job)
		return nil
	})

	// Register one or more topics and their processor
	m.Register("Plan", func(job *jobqueue.Job) error {
		time.Sleep(15 * time.Second)
		fmt.Println("Method1", job)
		return nil
	})
	
	time.Sleep(2 * time.Millisecond)

	statsReq := &jobqueue.StatsRequest{Topic: "Plan", CorrelationGroup: "2"}
	// Start the manager
	stats, err := m.Stats(context.Background(), statsReq)
	if err != nil {
		panic(err)
	}
	fmt.Println("Statistics", stats)

	if stats.Working == 0 && (stats.Succeeded + stats.Failed) == 0 {
		// Start the manager
		err = m.Start()
		if err != nil {
			panic(err)
		}
		fmt.Println("Worker Started")

		// Stop/Close the manager
		// err = m.Stop()
		// if err != nil {
		// 	fmt.Println("Stop failed")
		// 	return
		// }
		// fmt.Println("Stopped")
	}


	// Add a job: It'll be added to the store and processed eventually.
	// job := &jobqueue.Job{Topic: "clicks", Args: []interface{}{640, 480}}
	// err = m.Add(context.Background(), job)
	// if err != nil {
	// 	fmt.Println("Add failed")
	// 	return
	// }
	// fmt.Println("Job added")

	// // Stop/Close the manager
	// err = m.Stop()
	// if err != nil {
	// 	fmt.Println("Stop failed")
	// 	return
	// }
	// fmt.Println("Stopped")

}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{
		validator: validator.New(),
	}

	e.GET("/", func(c echo.Context) error {
		// params := c.QueryParams()
		// queries := c.QueryString()
		// pv := c.ParamValues()

		c.Response().Before(func() {
			println("before response")
		})

		c.Response().After(func() {
			println("after response")
		})

		queuejob()

		// return c.JSON(http.StatusOK, []interface{} {params, queries, pv})
		return c.NoContent(http.StatusNoContent)
	})

	e.POST("/users/:id/:test", getUser)
	e.GET("/users/:path", postUser)

	e.Logger.Fatal(e.Start("localhost:1323"))
}
