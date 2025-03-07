package handler

//Aqui estarnra los ruteos (osea defnir el path de endpount y los metodos (GET, POST) que usara, para manejarlo aqui en vez del main)
//Ademas tambine tendra manejo del middlware
//OSEA ESTARAN LOS RUTEOS Y LOS MIDDLWARE

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/IgnacioBO/go_lib_response/response"
	"github.com/IgnacioBO/gomicro_user/internal/user"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Este recibe contexto y un endpoint que definimos en la capa del endpionts
func NewUserHTTPServer(ctx context.Context, endpoints user.Endpoints) http.Handler {
	router := mux.NewRouter()

	//Middleware de gokit de server options
	//En caso de que haya un ERROR se pasa a una funcion creada por nsotros
	//Esta se guarad en opciones y se pone al final en Handle
	opciones := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	//Antiguo router.HandleFunc("/users", userEndpoint.Create).Methods("POST")
	//Ahora usaremos Handle, poreque a este se le puede pasar un server
	//Handre recibe un string y un http.Handler (una interfaz)
	//****Con httptranpsort podemos apsarle un server al Handle (NewServer devuelve un struct httptransport.Server)***
	//ese struct tiene una funncion -> func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request)
	// ESA FUNCION es la misma que usa que tiene la interfaz http.Handler y que se debe IMPLEMENTAR -> asi es -> type Handler interface {ServeHTTP(ResponseWriter, *Request)}
	// COMO ese struct (httptransport.Server) IMPLENENTA la funcion ServeHTTP de la interfaz http.Hnadler, puede pasarse aqui sin problemas
	router.Handle("/users", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Create), //EN el tutorial le pasa riecto endpoints.Create, PERO DEBIDO A MI VERSION MAS NUEV DE GO, se hace UNA CONVERSION de Type Controller a Endpoint de manera epxlicia (a pesar de q son identicads)
		decodeCreateUser,                    //le pasamos la funcion, como referencia para que la ocupe newServer
		encodeResponse,
		opciones...,
	)).Methods("POST")
	//Lo que hace lo de arriba es que cuando ahcemos peticion a /user, primer entra a la funicion Decode, el Decod le devuelve un STRUCT de request (que es de tipo CreatReq con fistname, lastname)
	//Ese struct del reques se lo envia el endpoints.Create -> luego va a servicion, repo, etc.
	//Al finalizar enpdoint.Create retorna el user.User (con struct con tag json, etc)
	//Luego el encode respone recibe ese user.User (resp interface{}) y luego se envia la respuesta

	router.Handle("/users/{id}", httptransport.NewServer( //Con httptranpsort podemos apsarle un server al HHandle
		endpoint.Endpoint(endpoints.Get), //Se va al Get del Endpoint
		decodeGetUser,                    //le pasamos la funcion decodeGetUser (distinta a la de create)
		encodeResponse,
		opciones...,
	)).Methods("GET")

	router.Handle("/users", httptransport.NewServer( //Con httptranpsort podemos apsarle un server al HHandle
		endpoint.Endpoint(endpoints.GetAll), //Se va al Get del Endpoint
		decodeGetAllUser,                    //le pasamos la funcion decodeGetAllUser (distinta a la de create)
		encodeResponse,
		opciones...,
	)).Methods("GET")

	router.Handle("/users/{id}", httptransport.NewServer( //
		endpoint.Endpoint(endpoints.Delete),
		decodeDeleteUser,
		encodeResponse,
		opciones...,
	)).Methods("DELETE")

	router.Handle("/users/{id}", httptransport.NewServer( //
		endpoint.Endpoint(endpoints.Update),
		decodeUpdateUser,
		encodeResponse,
		opciones...,
	)).Methods("PATCH")

	return router
}

// *** MIDDLEWARE REQUEST ***
// Funcion de decode, hacer decode dentro del REQUEST cuando usmoos Create d user, osea lo que aceoms anted en endpoint.go
// Devolver a una interface{} que en este caso sera un Struct del CreatReq (qeu tiene firstname, lastname, etc) y lo recibira el endpoint.go
func decodeCreateUser(_ context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de auth (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}
	var reqStruct user.CreateRequest

	//Ahora hacemos el decode del body del json al srtuct de REquest de usuario
	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}
	return reqStruct, nil

}

// *** MIDDLEWARE RESPONSE ***
// Esta funcion sera la que se necarge QUE DEVUEVLER EL ENDPOINT, osea esl a respiesta al cliente
// Oosea corre despues de endpoint.go (osea segun lo que devuelva la funcion con return aqui se maneja)
func encodeResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	rInterface := resp.(response.Response)                            //Transformamos el resp a response.Respone (al interface) -> YA QUE LE ENAIREMOS SIEMPRE UN objeto RESPONSE (CREADO POR NOSOTROS, q tiene el code, mensage, meta, etc, todo el json)
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	w.WriteHeader(rInterface.StatusCode())
	return json.NewEncoder(w).Encode(rInterface) //resp tendra el user.User del domain y otroas datos si es necesario para ocnveritse en json

}

// *** MIDDLEWARE RESPONSE DE ERROR ***
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	respInterface := err.(response.Response)                          //Tranfosrmamos el error recibido a la interfac response.Response que craemos
	//¿Porque funciona esta conversion de tipo error al de nosotros?, porque la interfaz 'error' de go pide que haya un metodo Error() string [QUE CREAMOS EN nuestro respon.RESPONSE!]
	//Entonces como implementamos el metodo Error() string funcinoa, ademas tenemos al ventaja que vamos apoder obtener MAS DATOS porque repsonse.Response tiene mas metodos como (StatusCode())
	//Entonces podemos transofrmar un error a una interfac propia con MAS METODOS Y MAS DATOS UE UN ERROR NORMAL!
	w.WriteHeader(respInterface.StatusCode())
	_ = json.NewEncoder(w).Encode(respInterface) //resp tendra el user.User del domain y otroas datos si es necesario para ocnveritse en json

}

// *** MIDDLEWARE REQUEST GET ***
// Funcion de decode, de GET
func decodeGetUser(_ context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de auth (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}
	var getReq user.GetRequest
	variablesPath := mux.Vars(r)
	getReq.ID = variablesPath["id"] //OBtenemos el id y lo guardamos en el cmapo ID de getReq

	fmt.Println("id es:", getReq.ID)

	return getReq, nil

}

// *** MIDDLEWARE REQUEST GET All ***
// Funcion de decode, de GET
func decodeGetAllUser(_ context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de auth (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}
	//Query() devielve un objeto que permite acceder a los parametros d la url (...?campo=123&campo2=hola)
	variablesURL := r.URL.Query()

	//Ahora obtendremos el limit y la pagina desde los parametros
	limit, _ := strconv.Atoi(variablesURL.Get("limit"))
	page, _ := strconv.Atoi(variablesURL.Get("page"))

	getReqAll := user.GetAllRequest{
		FirstName: variablesURL.Get("first_name"),
		LastName:  variablesURL.Get("last_name"),
		Limit:     limit,
		Page:      page,
	}

	return getReqAll, nil
}

// *** MIDDLEWARE REQUEST Delete ***
func decodeDeleteUser(_ context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de auth (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}
	variablesPath := mux.Vars(r)
	id := variablesPath["id"]
	fmt.Println("id a eliminar es:", id)
	deleteReq := user.DeleteRequest{ID: id}

	return deleteReq, nil

}

// *** MIDDLEWARE REQUEST Delete***
func decodeUpdateUser(_ context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de auth (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}
	var reqStruct user.UpdateRequest

	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}

	variablesPath := mux.Vars(r)
	reqStruct.ID = variablesPath["id"]

	return reqStruct, nil

}

// Authoruzation con tiken
func authorization(token string) error {
	if token != os.Getenv("TOKEN") {
		return errors.New("invalid token")
	}
	return nil
}
