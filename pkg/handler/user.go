package handler

//Aqui estarnra los ruteos (osea defnir el path de endpount y los metodos (GET, POST) que usara, para manejarlo aqui en vez del main)
//Ademas tambine tendra manejo del middlware
//OSEA ESTARAN LOS RUTEOS Y LOS MIDDLWARE

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IgnacioBO/gomicro_user/internal/user"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Este recibe contexto y un endpoint que definimos en la capa del endpionts
func NewUserHTTPServer(ctx context.Context, endpoints user.Endpoints) http.Handler {
	router := mux.NewRouter()

	//Antiguo router.HandleFunc("/users", userEndpoint.Create).Methods("POST")
	//Ahora usaremos Handle, poreque a este se le puede pasar un server
	router.Handle("/users", httptransport.NewServer( //Con httptranpsort podemos apsarle un server al HHandle
		endpoint.Endpoint(endpoints.Create), //EN el tutorial le pasa riecto endpoints.Create, PERO DEBIDO A MI VERSION MAS NUEV DE GO, se hace UNA CONVERSION de Type Controller a Endpoint de manera epxlicia (a pesar de q son identicads)
		decodeCreateUser,                    //le pasamos la funcion, como referencia para que la ocupe newServer
		encodeResponse,
	)).Methods("POST")
	//Lo que hace lo de arriba es que cuando ahcemos peticion a /user, primer entra a la funicion Decode, el Decod le devuelve un STRUCT de request (que es de tipo CreatReq con fistname, lastname)
	//Ese struct del reques se lo envia el endpoints.Create -> luego va a servicion, repo, etc.
	//Al finalizar enpdoint.Create retorna el user.User (con struct con tag json, etc)
	//Luego el encode respone recibe ese user.User (resp interface{}) y luego se envia la respuesta

	return router
}

// *** MIDDLEWARE REQUEST ***
// Funcion de decode, hacer decode dentro del REQUEST cuando usmoos Create d user, osea lo que aceoms anted en endpoint.go
// Devolver a una interface{} que en este caso sera un Struct del CreatReq (qeu tiene firstname, lastname, etc)
func decodeCreateUser(_ context.Context, r *http.Request) (interface{}, error) {
	var reqStruct user.CreateRequest

	//Ahora hacemos el decode del body del json al srtuct de REquest de usuario
	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, err
	}
	return reqStruct, nil

}

// *** MIDDLEWARE RESPONSE ***
// Esta funcion sera  la que se necarge QUE DEVUEVL EL ENDPOINT, osea esl a respiesta al cliente
func encodeResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	w.WriteHeader(200)
	return json.NewEncoder(w).Encode(resp) //resp tendra el user.User del domain y otroas datos si es necesario para ocnveritse en json

}
