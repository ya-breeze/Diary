import { HttpInterceptorFn } from "@angular/common/http";
import { inject } from "@angular/core";
import { AuthService } from "../services/auth.service";

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const token = authService.getToken();

  // Always include credentials (cookies) with requests
  // Backend must have proper CORS headers configured:
  // - Access-Control-Allow-Origin: <frontend-origin> (specific origin, not *)
  // - Access-Control-Allow-Credentials: true
  let clonedReq = req.clone({
    withCredentials: true,
  });

  // Skip adding token for login endpoint
  if (req.url.includes("/authorize")) {
    return next(clonedReq);
  }

  // Add the authorization header if token exists
  if (token) {
    clonedReq = clonedReq.clone({
      setHeaders: {
        Authorization: `Bearer ${token}`,
      },
    });
  }

  return next(clonedReq);
};
