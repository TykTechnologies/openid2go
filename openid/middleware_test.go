package openid

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

const idToken string = "IDTOKEN"

func Test_authenticateUser_WhenGetIDTokenReturnsError_WhenErrorHandlerContinues(t *testing.T) {
	_, c := createConfiguration(t, errorHandlerContinue, getIDTokenReturnsError)

	u, halt := authenticateUser(c, httptest.NewRecorder(), nil)

	if u != nil {
		t.Errorf("The returned user should be nil, but was %+v.", u)
	}

	if halt {
		t.Error("The authentication should have returned 'halt' false.")
	}
}

func Test_authenticateUser_WhenGetIDTokenReturnsError_WhenErrorHandlerHalts(t *testing.T) {
	_, c := createConfiguration(t, errorHandlerHalt, getIDTokenReturnsError)

	u, halt := authenticateUser(c, httptest.NewRecorder(), nil)

	if u != nil {
		t.Errorf("The returned user should be nil, but was %+v.", u)
	}

	if !halt {
		t.Error("The authentication should have returned 'halt' true.")
	}
}

func Test_authenticateUser_WhenValidateReturnsError_WhenErrorHandlerHalts(t *testing.T) {
	vm, c := createConfiguration(t, errorHandlerHalt, getIDTokenReturnsSuccess)
	go func() {
		vm.assertValidate(idToken, nil, errors.New("Error while validating the token."))
		vm.close()
	}()

	u, halt := authenticateUser(c, httptest.NewRecorder(), nil)

	if u != nil {
		t.Errorf("The returned user should be nil, but was %+v.", u)
	}

	if !halt {
		t.Error("The authentication should have returned 'halt' true.")
	}

	vm.assertDone()
}

func Test_authenticateUser_WhenValidateSucceeds(t *testing.T) {
	vm, c := createConfiguration(t, errorHandlerHalt, getIDTokenReturnsSuccess)
	iss := "https://issuer"
	sub := "SUB1"

	jt := &jwt.Token{Claims: jwt.MapClaims(make(map[string]interface{}))}
	jt.Claims.(jwt.MapClaims)["iss"] = iss
	jt.Claims.(jwt.MapClaims)["sub"] = sub

	go func() {
		vm.assertValidate(idToken, jt, nil)
		vm.close()
	}()

	u, halt := authenticateUser(c, httptest.NewRecorder(), nil)

	if halt {
		t.Error("A successful authenticateUser call should not have returned halt with value true.")
	}

	if u == nil {
		t.Fatal("The returned user should not be nil.")
	}

	if u.Issuer != iss {
		t.Error("Expected user issuer", iss, ", but got", u.Issuer)
	}

	if u.ID != sub {
		t.Error("Expected user ID", sub, ", but got", u.ID)
	}

	thisC := u.Claims
	thisJT := jt.Claims.(jwt.MapClaims)
	if len(thisC) != len(thisJT) {
		t.Error("Expected number of user claims", len(thisJT), ", but got", len(u.Claims))
	}

	vm.assertDone()
}

func createConfiguration(t *testing.T, eh ErrorHandlerFunc, gt GetIDTokenFunc) (*jwtTokenValidatorMock, *Configuration) {
	jm := newJwtTokenValidatorMock(t)
	c, _ := NewConfiguration(ErrorHandler(eh))
	c.tokenValidator = jm
	c.IDTokenGetter = gt
	return jm, c
}

func getIDTokenReturnsError(r *http.Request) (string, error) {
	return "", errors.New("An error happened when returning ID Token.")
}

func getIDTokenReturnsSuccess(r *http.Request) (string, error) {
	return idToken, nil
}

func errorHandlerHalt(e error, w http.ResponseWriter, r *http.Request) bool {
	if e != nil {
		return true
	}
	return false
}

func errorHandlerContinue(e error, w http.ResponseWriter, r *http.Request) bool {
	return false
}
