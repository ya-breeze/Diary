import { inject } from "@angular/core";
import { Router, CanActivateFn } from "@angular/router";
import { map } from "rxjs";
import { AuthService } from "../services/auth.service";

export const authGuard: CanActivateFn = (route, state) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  // Validate authentication by checking token and loading user profile if needed
  return authService.validateAuthentication().pipe(
    map((isAuthenticated) => {
      if (isAuthenticated) {
        return true;
      }

      // Redirect to login page with return url
      router.navigate(["/login"], { queryParams: { returnUrl: state.url } });
      return false;
    })
  );
};
