package user

//**Capa endpoint o controlador**

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/IgnacioBO/gomicro_meta/meta"
	"github.com/gorilla/mux"
)

// Struct que tenga todos los endpoints que vayamos a utilizar
// Que teng una fucion que recibe un request y un response
type (
	//Controller sera una funcion que reciba REspone y Request
	Controller func(w http.ResponseWriter, r *http.Request)
	Endpoints  struct {
		Create        Controller //Esto es lo mismo que decir Create func(w http.ResponseWriter, r *http.Request), pero como TODOS SON tipo Controller (Definido arriba) nos ahorramos ahcerlo
		Get           Controller
		GetAll        Controller
		Update        Controller
		Delete        Controller
		DeleteClassic Controller
	}
	//Definiremos una struct para definir el request del Craete, con los campos que quiero recibir y los tags de json
	CreateRequest struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	}
	//Definiremos una struct para definir el request del UPDATE, con los campos que quiero y SE PODRAN ACTUALIZAR y los tags de json
	//Seran de tipo puntero * para que puedan venir vacios y poder separar entre vacios "" y que no vengan
	UpdateRequest struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Email     *string `json:"email"`
		Phone     *string `json:"phone"`
	}

	MsgResponse struct {
		ID  string `json:"id"`
		Msg string `json:"msg"`
	}

	//Este sera el "response" generico, para tener una estructura
	//El status SIEMPRE sera devuelto
	//El campo data SOLO aparecera cuando esta todo OK y dentro ira la estructura
	//El campo error SOLO aparecera cuando hay error
	Response struct {
		Status int         `json:"status"`
		Data   interface{} `json:"data,omitempty"` //omitempty, asi cuando queremos enviamos la data cuando eta ok y cuando este eror se envie el campo error
		Err    string      `json:"error,omitempty"`
		Meta   *meta.Meta  `json:"meta,omitempty"`
	}

	//Struct para guardar la cant page por defecto y otras conf
	Config struct {
		LimitPageDefault string
	}
)

// Funcion que se encargará de hacer los endopints
// Para eso necesitaremos una struct que se llamara endpoints
// Esta funcion va a DEVOLVER una struct de Endpoints, estos endpoints son los que vamos a poder utuaizlar en unestro dominio (user)
func MakeEndpoints(s Service, config Config) Endpoints {
	return Endpoints{
		Create: makeCreateEndpoint(s),
		Get:    makeGetEndpoint(s),
		Update: makeUpdateEndpoint(s),
		Delete: makeDeleteEndpoint(s),
		GetAll: makeGetAllEndpoint(s, config),
	}
}

// Este devolver un Controller, retora una función de tipo Controller (que definimos arriba) con esta caractesitica
// Es privado porque se llamar solo de este dominio
func makeDeleteEndpoint(s Service) Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("delete user")
		w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json

		variablesPath := mux.Vars(r)
		id := variablesPath["id"]
		fmt.Println("id a eliminar es:", id)
		err := s.Delete(id)
		if err != nil {
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(&Response{Status: 404, Err: err.Error()}) //Aqui devolvemo el posible erro
			return
		}

		//Aqui le pasamos Response como struct para ahorrar memoria
		//Con puntero (&Response): Encode accede al struct original a través de la dirección de memoria. Esto evita copiar los datos.
		//Sin puntero (Response): Encode recibe una copia del struct completo, lo que puede ocupar más memoria si el struct es muy grande.
		json.NewEncoder(w).Encode(&Response{Status: 200, Data: map[string]string{"id": id, "msg": "success"}})
	}
}

// w http.ResponseWriter -> Para enviar RESPUESTA AL CLIENTE (body, headers, statsu code)
// r *http.Request -> Contiende info de la SOLICITUD/REQUEST del cliente (aaceder al metodo http (GET,POST,ETC), paarametros url/query string, body, headers, etc
// *http.Request siempre como PUNTERO (*) por mas eficiencia, para poder modificar datos y es el estandar del package net/http
func makeCreateEndpoint(s Service) Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("create user")
		w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json

		//Variable con struct de request (datos usaurio)
		var reqStruct CreateRequest
		//r.Body tiene el body del request (se espera JSON) y lo decodifica al struct (reqStruct) (osea pasar el json enviado en el request a un struct)
		err := json.NewDecoder(r.Body).Decode(&reqStruct)
		if err != nil {
			//w.WriteHeader devuelve en el repsonse el CODE que se le indica
			w.WriteHeader(400)
			//Enviaremos la repsuesta con encode y creamos un Sruct ErrorRespone (Creado antes) con un texto
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: "invalid request format"})

			json.NewEncoder(w).Encode(&Response{Status: 400, Err: "invalid request format"})
			return
		}

		//Validaciones
		if reqStruct.FirstName == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: "first_name is required"})
			return
		}
		if reqStruct.LastName == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: "last_name is required"})
			return
		}
		fmt.Println(reqStruct)
		reqStrucEnJson, _ := json.MarshalIndent(reqStruct, "", " ")
		fmt.Println(string(reqStrucEnJson))

		//Usaremos la s recibida como parametro (de la capa Service y usaremos el metodo CREATE con lo que debe recibir)
		usuarioNuevo, err := s.Create(reqStruct.FirstName, reqStruct.LastName, reqStruct.Email, reqStruct.Phone)
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: err.Error()}) //Aqui devolvemo el posible erro
			return
		}

		//Para responder se usa el paquete json Encode (devolverá en el response(w) un JSON, este JSON será la transformacion del struct en json usando la funcion Encode)
		//Antes devolviemoas el reqStruct (que ERA LO MISMO QUE ENVIA EL CLIENTE)
		//Pero ahora devolveremos usuarioNuevo que seria el struct User (del dominio) que tiene como se inserto a la BBDD
		json.NewEncoder(w).Encode(&Response{Status: 200, Data: usuarioNuevo})
	}
}

func makeUpdateEndpoint(s Service) Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("update user")
		w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json

		//Variable con struct de request (datos de atualizacion)
		var reqStruct UpdateRequest
		//r.Body tiene el body del request (se espera JSON) y lo decodifica al struct (reqStruct) (osea pasar el json enviado en el request a un struct)
		err := json.NewDecoder(r.Body).Decode(&reqStruct)
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: "invalid request format"})
			return
		}

		//Validaciones para que sean reqierod
		//Si first name es disinto de nil (osea el puntero NO VIENE VACIO) y le pone "first_name" como vacio (osea el cliene pone first_name = "") da error
		//PERO SI EL CLIENTE NO ENVIA first_namem reqStruct.Firstname sera igual a NIL! entonces no entra
		//OSea se permite NO ENVIAR ESTOS CAMPOS, PERO NO SE PERMITE ENVIARLSO VACIOS
		if reqStruct.FirstName != nil && *reqStruct.FirstName == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: "first_name can't be empty"})
			return
		}

		if reqStruct.LastName != nil && *reqStruct.LastName == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: "last_name can't be empty"})
			return
		}
		variablesPath := mux.Vars(r)
		id := variablesPath["id"]

		err = s.Update(id, reqStruct.FirstName, reqStruct.LastName, reqStruct.Email, reqStruct.Phone)
		if err != nil {
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(Response{Status: 404, Err: err.Error()}) //Aqui devolvemo el posible erro
			return
		}
		json.NewEncoder(w).Encode(&Response{Status: 200, Data: map[string]string{"id": id, "msg": "success"}})

	}
}
func makeGetEndpoint(s Service) Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("get user")
		w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json

		//Aqui usamos MUX para extrar las variables del path (url)
		//Y con ["id"] obtenemos el valor del parametro {id} definido en el main.go (users/{id})
		//**¿ESTA BIEN ESTO?? USAMOS LIBRERIA EXTERNA EN "internal/users", no se deberia -> en notasAparte.txt tengo una solucion
		variablesPath := mux.Vars(r)
		id := variablesPath["id"]
		fmt.Println("id es:", id)
		usuario, err := s.Get(id)
		if err != nil {
			if usuario == nil { //Si usuario es vacio da 404
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(&Response{Status: 404, Err: err.Error() + ". user with id " + id + " doesn't exist"}) //Aqui devolvemo el posible erro
				return
			} else {
				w.WriteHeader(400)
				json.NewEncoder(w).Encode(&Response{Status: 400, Err: err.Error()}) //Aqui devolvemo el posible erro
				return
			}
		}

		json.NewEncoder(w).Encode(&Response{Status: 200, Data: usuario})

	}
}
func makeGetAllEndpoint(s Service, config Config) Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("getall user")
		w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json

		//URL.Query() que viene del request
		//Query() devielve un objeto que permite acceder a los parametros d la url (...?campo=123&campo2=hola)
		variablesURL := r.URL.Query()
		//Luego con podemos acceder a los parametos y guardarlos en el struct Filtro (creado en service.go)
		filtros := Filtros{
			FirstName: variablesURL.Get("first_name"),
			LastName:  variablesURL.Get("last_name"),
		}
		//Ahora obtendremos el limit y la pagina desde los parametros
		limit, _ := strconv.Atoi(variablesURL.Get("limit"))
		page, _ := strconv.Atoi(variablesURL.Get("page"))

		//Ahora llamaremos al Count del service que creamos (antes de hacer la consulta completa)
		cantidad, err := s.Count(filtros)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(&Response{Status: 500, Err: err.Error()}) //Aqui devolvemo el posible erro
			return
		}
		//Luego crearemos un meta y le agregaremos la cantidad que consultamos, luego el meta lo ageregaremos a la respuesta
		meta, err := meta.New(page, limit, cantidad, config.LimitPageDefault)

		allUsers, err := s.GetAll(filtros, meta.Offset(), meta.Limit()) //GetAll recibe el offset (desde q resultado mostrar) y el limit (cuantos desde el offset)
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&Response{Status: 400, Err: err.Error()}) //Aqui devolvemo el posible erro
			return
		}

		json.NewEncoder(w).Encode(&Response{Status: 200, Data: allUsers, Meta: meta})
	}
}
