import SendRequest from "./xhr";

export const GetAuthor = ({ id, slug, name }: GetEntityOptions) =>
  SendRequest<Author>("GET", `/api/author?id=${id || ""}&slug=${slug || ""}&name=${name || ""}`);

export const CreateAuthor = (name: string) => SendRequest<Author>("POST", "/api/author", JSON.stringify({ name }));

export const DeleteAuthor = (slugOrName: string) => SendRequest("DELETE", `/api/author/${slugOrName}`);

export const GetAuthors = () => SendRequest<Author[]>("GET", "/api/authors");
