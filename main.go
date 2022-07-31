package main

import (
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"rxdrag.com/entify/config"
	"rxdrag.com/entify/consts"
	"rxdrag.com/entify/db"
	"rxdrag.com/entify/handler"
	"rxdrag.com/entify/middlewares"
	"rxdrag.com/entify/repository"
	"rxdrag.com/entify/resolve"
	"rxdrag.com/entify/schema"
)

const PORT = 4000

func checkParams() {
	dbConfig := config.GetDbConfig()
	if dbConfig.Driver == "" ||
		dbConfig.Host == "" ||
		dbConfig.Database == "" ||
		dbConfig.User == "" ||
		dbConfig.Port == "" ||
		dbConfig.Password == "" {
		panic("Params is not enough, please set")
	}
}

func checkMetaInstall() {
	if !repository.IsEntityExists(consts.META_ENTITY_NAME) {
		schema.Installed = false
	} else {
		schema.Installed = true
	}
}

func main() {
	defer db.Close()
	checkParams()
	checkMetaInstall()

	h := handler.New(&handler.Config{
		Pretty:         true,
		GraphiQLConfig: &handler.GraphiQLConfig{},
	})

	http.Handle("/graphql",
		middlewares.CorsMiddleware(
			middlewares.ContextMiddleware(
				resolve.LoadersMiddleware(h),
			),
		),
	)
	if config.Storage() == consts.LOCAL {
		fmt.Println(fmt.Sprintf("Running a file server at http://localhost:%d/uploads/", PORT))
		http.Handle(consts.UPLOAD_PRIFIX+"/", http.StripPrefix(consts.UPLOAD_PRIFIX, http.FileServer(http.Dir(consts.UPLOAD_PATH))))
	}

	fmt.Println(fmt.Sprintf("🚀 Graphql server ready at http://localhost:%d/graphql", PORT))
	err2 := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil)
	if err2 != nil {
		fmt.Printf("启动失败:%s", err2)
	}
}
