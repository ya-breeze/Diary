import { Routes } from "@angular/router";
import { authGuard } from "./core/guards/auth.guard";
import { LoginComponent } from "./auth/login/login.component";
import { DiaryListComponent } from "./diary/diary-list/diary-list.component";
import { DiaryEditorComponent } from "./diary/diary-editor/diary-editor.component";
import { DiarySearchComponent } from "./diary/diary-search/diary-search.component";

export const routes: Routes = [
  { path: "", redirectTo: "/diary", pathMatch: "full" },
  { path: "login", component: LoginComponent },
  {
    path: "diary",
    component: DiaryEditorComponent,
    canActivate: [authGuard],
  },
  {
    path: "diary/list",
    component: DiaryListComponent,
    canActivate: [authGuard],
  },
  {
    path: "diary/search",
    component: DiarySearchComponent,
    canActivate: [authGuard],
  },
  { path: "**", redirectTo: "/diary" },
];
