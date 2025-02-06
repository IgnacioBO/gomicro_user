package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IgnacioBO/gomicro_user/internal/user"

	"github.com/IgnacioBO/gomicro_user/pkg/bootstrap"
	"github.com/gorilla/mux" //Manejar ruteo facilmente (paths y metodos)
	"github.com/joho/godotenv"
	//Driver mysql para gorm
	//Para manejar bbdd facilmente con strcut y funciones (en vez de querys directa)
)

func main() {
	//Generaremos un router usando gorilla/mux para generar un RUTEO (osea los paths y metodos)
	router := mux.NewRouter()

	//Definiremos un logger
	//Creamos un script bootsrap.go que INICIALIZA el log
	l := bootstrap.InitLogger()

	//Con godotenv cargamos las variables de entorn denrto de .env para usarlas en el DSN
	//Al usar godotenv.Load() cargar autoamticametne los valores en archivo .env de root
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	//Creamos un config que tenga la cant max de pagina por defecto
	pageLimDef := os.Getenv("PAGINATOR_LIMIT_DEFAULT")
	if pageLimDef == "" {
		l.Fatal("paginator limit default is required")
	}
	userConfig := user.Config{LimitPageDefault: pageLimDef}

	//DSN (Data Source Name) es una cadena de conexion de BBDD (tipo, servidor, nombre bbdd, user, pass)
	//Creamos un script bootsrap.go que INICIALIZA la conexion usando gorm y varaibles de entoero
	db, err := bootstrap.DBConnection()
	if err != nil {
		log.Fatal(err)
	}

	//Generaremos un objeto repo (que recibe la bbdd y logger) que luego le pasaremos a la capa servicio
	userRepo := user.NewRepo(l, db)
	//Crearemos un objeto de tipo servicio pasandole un objeto Repository (y logger) para luego pasarselo a la capa enpdoint
	userService := user.NewService(l, userRepo)
	//Crearemo un objeto de tipo endpoint y le pasamos el objeto creado (Service). Ademas le pasamos un user.Config
	userEndpoint := user.MakeEndpoints(userService, userConfig)

	//Ahora setearemos que cuando le pegemos a /users le pege a las funciones definidas en el controlador user
	//Con handlefunc decimos que cuando valla a /users se ejecute la funcion correspondiente (userEnd.Create, Get, etc)
	//Podemos PONER y ESPECIFICAR EL METODO (si se quiere), si intento pegarle con otro no soportado tirará error 405
	router.HandleFunc("/users", userEndpoint.Create).Methods("POST")
	router.HandleFunc("/users", userEndpoint.GetAll).Methods("GET")
	router.HandleFunc("/users/{id}", userEndpoint.Update).Methods("PATCH")  //Usamos {} para especifiar el NOMBRE del paramatro (que se obtiene con MUX dentro de endpoint.go) //Patch es ACT parcial (PUT es completa)
	router.HandleFunc("/users/{id}", userEndpoint.Delete).Methods("DELETE") //Delete o SoftDelete
	router.HandleFunc("/users/{id}", userEndpoint.Get).Methods("GET")       //Usamos {} para especifiar el NOMBRE del paramatro (que se obtiene con MUX dentro de endpoint.go)

	port := os.Getenv("PORT")
	address := fmt.Sprintf("127.0.0.1:%s", port)

	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		ReadTimeout:  5 * time.Second, //Con estos SETEAMOS TIMEOUT DE ESCRITURA Y DE LECTURA (cuanto timepo maximo la api permite)
		WriteTimeout: 5 * time.Second, // Read es REQUEST, WRITE es RESPONE
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
