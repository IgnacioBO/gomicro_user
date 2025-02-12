package user

import (
	"errors"
	"fmt"
)

//Aqui definiermos errores ESTANDAR para usarlos (y evitar repetir el error)

var ErrFirstNameRequired = errors.New("first_name is required")
var ErrLastNameRequired = errors.New("last_name is required")

var ErrFirstNameNotEmpty = errors.New("first_name can't be empty")
var ErrLastNameNotEmpty = errors.New("last_name can't be empty")

// Revsiand oe lapckae errors de go nos daos cuento que el error es un struct que tiene UN STRING y la funcion "Error()" deielve ese string
// Adaptreamso esta logica para que podemos parametrizar el error
type ErrUserNotFound struct {
	UserID string
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user with id: %s not found", e.UserID)
}

//Mi solucion USABA una FUNCION y servia igual o simiar, a la funcion se le pasa ErrUserNotFound y se compara luego con error.Is
//var ErrUserNotFound = errors.New("user not found")

/*
func ErrUserNotFoundID(id string) error {
	return fmt.Errorf("user with id: %s not found: %w", id, ErrUserNotFound)
}*/
