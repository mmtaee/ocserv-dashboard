package middlewares

//func extractSection(path string) (string, bool) {
//	const prefix = "/api/"
//	if !strings.HasPrefix(path, prefix) {
//		return "", false
//	}
//	trimmed := strings.TrimPrefix(path, prefix)
//	parts := strings.SplitN(trimmed, "/", 2)
//	if len(parts) > 0 {
//		return parts[0], true
//	}
//	return "", false
//}
//
//func toMap(data interface{}) map[string]interface{} {
//	b, _ := json.Marshal(&data)
//	var dataStruct map[string]interface{}
//	_ = json.Unmarshal(b, &dataStruct)
//	return dataStruct
//}

//func RoutePermission() echo.MiddlewareFunc {
//	return func(next echo.HandlerFunc) echo.HandlerFunc {
//		return func(c echo.Context) error {
//			if c.Get("isAdmin").(bool) {
//				return next(c)
//			}
//
//			section, ok := extractSection(c.Path())
//			if !ok {
//				return PermissionDeniedError(c, "url section permission denied")
//			}
//
//			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//			defer cancel()
//			db := database.Get()
//			uid := c.Get("userUID").(string)
//			user := &models.User{}
//			if err := db.WithContext(ctx).Model(&models.User{UID: uid}).Where("uid = ? ", uid).First(&user).Error; err != nil {
//				return PermissionDeniedError(c, "user with permission not exist")
//			}
//
//			mapPermission := toMap(user.Permissions)
//			if value, found := mapPermission[section]; found {
//				if !value.(bool) {
//					return PermissionDeniedError(c, "you don't have permission to access this route")
//				}
//			} else {
//				return PermissionDeniedError(c, "permission not found")
//			}
//			return next(c)
//		}
//	}
//}
