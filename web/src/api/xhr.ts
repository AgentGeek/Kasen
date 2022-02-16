const ParseResult = <T = void>(xhr: XMLHttpRequest): ApiResult<T> => {
  let response: T;
  let error: ApiError;

  if (xhr.responseText) {
    if (xhr.status >= 200 && xhr.status < 300) {
      response = JSON.parse(xhr.responseText);
    } else if (xhr.status) {
      error = JSON.parse(xhr.responseText).error;
    }
  } else if (!xhr.status) {
    error = {
      message: "Unable to send request",
      cause: "ERR_INTERNET_DISCONNECTED"
    };
  }
  return { response, error, status: xhr.status };
};

const xhrEvents = ["abort", "error", "load", "loadend", "loadstart", "progress", "timeout"];
const SendRequest = <T = void>(
  method: string,
  url: string,
  data: Document | XMLHttpRequestBodyInit = undefined
): Api<T> => {
  const xhr = new XMLHttpRequest();
  xhr.open(method, url);

  const promise = new Promise<ApiResult<T>>((resolve, reject) => {
    xhr.addEventListener("readystatechange", () => {
      if (xhr.readyState !== XMLHttpRequest.DONE) return;
      resolve(ParseResult(xhr));
    });
    xhr.addEventListener("abort", () => reject());
  });

  const eventListeners: { [k: string]: EventListener[] } = {};
  const uploadEventListeners: { [k: string]: EventListener[] } = {};

  xhrEvents.forEach(k => {
    xhr.addEventListener(k, e => eventListeners[k]?.forEach(fn => window.setTimeout(() => fn(e), 0)));
    xhr.upload.addEventListener(k, e => uploadEventListeners[k]?.forEach(fn => window.setTimeout(() => fn(e), 0)));
  });
  xhr.send(data);

  return Object.assign(promise, {
    addEventListener(t, fn) {
      (eventListeners[t] ??= []).push(fn);
    },
    upload: {
      addEventListener(t, fn) {
        (uploadEventListeners[t] ??= []).push(fn);
      }
    },
    cancel() {
      xhr.abort();
    }
  });
};

export default SendRequest;
