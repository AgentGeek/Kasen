import deepEqual from "fast-deep-equal";
import React, { useEffect, useRef } from "react";
import { GetServiceConfig, UpdateServiceConfig } from "../../../api/config";
import { createRenderer } from "../../Renderer";
import Spinner, { WithSpinner } from "../../Spinner";
import { useToast } from "../../Toast";

const Service = () => {
  const toast = useToast();
  const mutex = useRef(false);
  const render = createRenderer();

  const configRef = useRef<ServiceConfig>();
  const isLoadingRef = useRef(true);

  useEffect(() => {
    GetServiceConfig().then(({ response, error }) => {
      if (response) configRef.current = response;
      if (error) toast.showError(error);

      isLoadingRef.current = false;
      render();
    });
  }, []);

  const isSubmittingRef = useRef<Dispatcher<boolean>>();
  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const target = e.target as HTMLFormElement;

    const config = {
      disableRegistration: target.disableRegistration.checked,
      coverMaxFileSize: Number(target.coverMaxFileSize.value),
      pageMaxFileSize: Number(target.pageMaxFileSize.value)
    };

    if (mutex.current || deepEqual(config, configRef.current)) {
      return;
    }
    mutex.current = true;
    isSubmittingRef.current(true);

    await (async () => {
      const { error, status } = await UpdateServiceConfig(config);
      if (status === 204) {
        toast.show("Service config has been updated");

        configRef.current = config;
        render();
      }
      if (error) toast.showError(error);
    })();

    mutex.current = false;
    isSubmittingRef.current(false);
  };

  return (
    <section className="service">
      <h3>Service</h3>
      {isLoadingRef.current ? (
        <Spinner width="60" height="60" strokeWidth="8" />
      ) : (
        <form onSubmit={onSubmit}>
          <div className="checkboxContainer">
            <strong>Disable Registration</strong>
            <input
              className="checkbox"
              type="checkbox"
              name="disableRegistration"
              defaultChecked={configRef.current.disableRegistration}
              placeholder="Required"
            />
          </div>
          <div className="inputContainer">
            <strong>Cover Max. File Size</strong>
            <input
              className="input"
              type="number"
              name="coverMaxFileSize"
              defaultValue={configRef.current.coverMaxFileSize}
              placeholder="Required (in bytes)"
              required
            />
          </div>
          <div className="inputContainer">
            <strong>Page Max. File Size</strong>
            <input
              className="input"
              type="number"
              name="pageMaxFileSize"
              defaultValue={configRef.current.pageMaxFileSize}
              placeholder="Required (in bytes)"
              required
            />
          </div>
          <button className="button green" type="submit">
            <WithSpinner width="16" height="16" strokeWidth="8" dispatcherRef={isSubmittingRef} />
            <strong>Update Service</strong>
          </button>
        </form>
      )}
    </section>
  );
};

export default Service;
