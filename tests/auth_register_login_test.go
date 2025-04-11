package tests

import (
	"sso/tests/suite"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/salta0/sso-protos/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})

	loginTime := time.Now()
	token := respLogin.GetToken()

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReg.GetId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegister_DuplicateRegister(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetId())

	respRegDup, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, respRegDup.GetId())
	assert.ErrorContains(t, err, "user already registred")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	cases := []struct {
		name     string
		email    string
		password string
		expError string
	}{
		{
			name:     "Registr with empty password",
			email:    gofakeit.Email(),
			password: "",
			expError: "password is required",
		},
		{
			name:     "Register with empty email",
			email:    "",
			password: randomFakePassword(),
			expError: "email is required",
		},
		{
			name:     "Email and password are empty",
			email:    "",
			password: "",
			expError: "email is required",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email: c.email, Password: c.password,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, c.expError)
		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	password := randomFakePassword()

	cases := []struct {
		name     string
		email    string
		password string
		app_id   int32
		expError string
	}{
		{
			name:     "Login with empty password",
			email:    email,
			password: "",
			app_id:   1,
			expError: "password is required",
		},
		{
			name:     "Login with empty email",
			email:    "",
			password: password,
			app_id:   1,
			expError: "email is required",
		},
		{
			name:     "Email and login are empty",
			email:    "",
			password: "",
			app_id:   1,
			expError: "email is required",
		},
		{
			name:     "Wrong email",
			email:    "wrong" + email,
			password: password,
			app_id:   1,
			expError: "invalid login or password",
		},
		{
			name:     "Wrong password",
			email:    email,
			password: "wrong" + password,
			app_id:   1,
			expError: "invalid login or password",
		},
		{
			name:     "Wrong email and password",
			email:    "wrong" + email,
			password: "wrong" + password,
			app_id:   1,
			expError: "invalid login or password",
		},
		{
			name:     "Invalid app id",
			email:    email,
			password: password,
			app_id:   0,
			expError: "app_id is required",
		},
	}

	registerResp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email: email, Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, registerResp)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			loginResp, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email: c.email, Password: c.password, AppId: c.app_id,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, c.expError)
			assert.Empty(t, loginResp)
		})
	}
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
