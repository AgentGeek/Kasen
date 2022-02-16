import SendRequest from "./xhr";

export const DeleteUser = (password: string) => SendRequest("DELETE", "/api/user", JSON.stringify({ password }));

export const DeleteUserById = (id: number) => SendRequest("DELETE", `/api/user/${id}`);

export const GetUser = () => SendRequest<User>("GET", "/api/user");

export const GetUsers = () => SendRequest<User[]>("GET", "/api/users");

export const UpdateUserName = (name: string) => SendRequest("PATCH", "/api/user/name", JSON.stringify({ name }));

export const UpdateUserNameById = (id: number, name: string) =>
  SendRequest("PATCH", `/api/user/${id}/name`, JSON.stringify({ name }));

export const UpdateUserPasssword = (currentPassword: string, newPassword: string) =>
  SendRequest("PATCH", "/api/user/password", JSON.stringify({ currentPassword, newPassword }));

export const UpdateUserPassswordById = (id: number, currentPassword: string, newPassword: string) =>
  SendRequest("PATCH", `/api/user/${id}/password`, JSON.stringify({ currentPassword, newPassword }));

export const UpdateUserPermissions = (id: number, ...permissions: string[]) =>
  SendRequest("PATCH", `/api/user/${id}/permissions`, JSON.stringify({ permissions }));
