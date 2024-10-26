package middlewares

import (
	_ "google.golang.org/api/idtoken"
	_ "google.golang.org/grpc/metadata"
)

//const (
//	googleClientID     = "153032778434-0i9hj6vckoelnv9c4sbb5861fujihajb.apps.googleusercontent.com"
//	googleClientSecret = ""
//	redirectURI        = "your_redirect_uri"
//)
//
//var (
//	AnonymousAPI = []string{
//		"/platform.v1.general.GeneralService/CreateJWTToken",
//	}
//)
//
//// JWTAuthMiddleware 认证中间件
//func JWTAuthMiddleware(ctx context.Context) (context.Context, error) {
//	tokens := metadata.ValueFromIncomingContext(ctx, strings.ToLower("authorization"))
//	if len(tokens) == 0 {
//		return nil, fmt.Errorf("no token found")
//	}
//
//	llJwtClaims, err := lbjwt.TokenToJwtClaims(tokens[0])
//	if err != nil {
//		return nil, err
//	}
//	currentUser := &user.User{
//		UserEmail: llJwtClaims.UserEmail,
//	}
//	return user.ToContext(ctx, currentUser), nil
//
//}
//
//type GoogleTokenInfo struct {
//	Audience      string      `json:"aud"`
//	Email         string      `json:"email"`
//	EmailVerified string      `json:"email_verified"`
//	ExpiresAt     json.Number `json:"exp"`
//	IssuedAt      string      `json:"iat"`
//	Issuer        string      `json:"iss"`
//	Name          string      `json:"name"`
//	Picture       string      `json:"picture"`
//	GivenName     string      `json:"given_name"`
//	FamilyName    string      `json:"family_name"`
//	Locale        string      `json:"locale"`
//}
//
//func GoogleOAuth2AuthMiddleware(ctx context.Context) (context.Context, error) {
//	tokens := metadata.ValueFromIncomingContext(ctx, strings.ToLower("authorization"))
//	if len(tokens) == 0 {
//		return nil, fmt.Errorf("no token found")
//	}
//	accessToken := tokens[0]
//
//	url := "https://www.googleapis.com/oauth2/v3/tokeninfo?access_token=" + accessToken
//	response, err := http.Get(url)
//	if err != nil {
//		return nil, err
//	}
//	defer response.Body.Close()
//
//	// 检查响应状态码
//	if response.StatusCode != http.StatusOK {
//		log.Errorf("Invalid response status code from Google OAuth: %v", response.Body)
//		return nil, fmt.Errorf("Invalid response status code from Google OAuth: %d", response.StatusCode)
//	}
//
//	// 解析响应
//	tokenInfo := GoogleTokenInfo{}
//	if err := json.NewDecoder(response.Body).Decode(&tokenInfo); err != nil {
//		return nil, err
//	}
//
//	expTimeUnix, err := tokenInfo.ExpiresAt.Int64()
//	if err != nil {
//		return nil, err
//	}
//	expTime := time.Unix(expTimeUnix, 0)
//	if time.Now().After(expTime) {
//		return nil, fmt.Errorf("Google Token expired")
//	}
//	if tokenInfo.Audience != googleClientID {
//		return nil, fmt.Errorf("Invalid audience")
//	}
//	// put user info to context
//	userInfo := &user.User{
//		UserEmail: tokenInfo.Email,
//	}
//	ctx = user.ToContext(ctx, userInfo)
//	return ctx, nil
//}
//
//// nolint
//func validateIdToken(ctx context.Context) (context.Context, error) {
//	token := "authorization token"
//	payload, err := idtoken.Validate(context.Background(), token, "your client id")
//	if err != nil {
//		log..GetDefaultLightboxLogger().Error(err.Error())
//	}
//	log.Infof("payload: %+v", payload)
//	return ctx, nil
//}
