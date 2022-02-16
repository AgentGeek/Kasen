import React, { useContext, useRef } from "react";
import { AlertTriangle } from "react-feather";
import { DeleteUser, UpdateUserName, UpdateUserPasssword } from "../../../api";
import { Permission } from "../../../constants";
import { HasPerms } from "../../../utils/utils";
import { createRenderer } from "../../Renderer";
import { WithSpinner } from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";

const User = () => {
  const { user } = useContext(ManageContext);
  const toast = useToast();

  const mutex = useRef(false);
  const render = createRenderer();

  const setIsUpdatingProfileRef = useRef<Dispatcher<boolean>>();
  const onProfileSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const name = (e.target as HTMLFormElement).username.value.trim();
    if (mutex.current || name === user.name) {
      return;
    }
    mutex.current = true;

    await (async () => {
      if (name.length < 3) {
        toast.show({
          type: "error",
          content: "User name must be at least 3 characters"
        });
        return;
      }

      if (name.length > 32) {
        toast.show({
          type: "error",
          content: "User name must be at most 32 characters"
        });
        return;
      }

      setIsUpdatingProfileRef.current(true);
      const { error, status } = await UpdateUserName(name);
      if (status === 204) {
        toast.show("User name has been updated");
        user.name = name;

        render();
      }
      if (error) toast.showError(error);
      setIsUpdatingProfileRef.current(false);
    })();

    mutex.current = false;
  };

  const setIsUpdatingPasswordRef = useRef<Dispatcher<boolean>>();
  const onPasswordSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (mutex.current) return;
    mutex.current = true;

    const target = e.target as HTMLFormElement;
    const currentPassword = target.currentPassword.value;
    const newPassword = target.newPassword.value;

    await (async () => {
      if (currentPassword.length < 6) {
        toast.show({
          type: "error",
          content: "Current password must be at least 6 characters"
        });
        return;
      }

      if (newPassword.length < 6) {
        toast.show({
          type: "error",
          content: "New password must be at least 6 characters"
        });
        return;
      }

      setIsUpdatingPasswordRef.current(true);
      const { error, status } = await UpdateUserPasssword(currentPassword, newPassword);
      if (status === 204) {
        toast.show("Password has been updated");

        target.currentPassword.value = "";
        target.newPassword.value = "";
        render();
      }
      if (error) toast.showError(error);
      setIsUpdatingPasswordRef.current(false);
    })();

    mutex.current = false;
  };

  const setIsDeletingRef = useRef<Dispatcher<boolean>>();
  const onDeleteAccountSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (mutex.current) return;
    mutex.current = true;

    const target = e.target as HTMLFormElement;
    const password = target.password.value;

    await (async () => {
      if (password.length < 6) {
        toast.show({
          type: "error",
          content: "Password must be at least 6 characters"
        });
        return;
      }

      setIsDeletingRef.current(true);
      const { status, error } = await DeleteUser(password);

      if (status === 204) {
        toast.show("Account has been deleted");

        setTimeout(() => {
          window.location.href = "/";
        }, 3000);
      }
      if (error) toast.showError(error);

      setIsDeletingRef.current(false);
    })();

    mutex.current = false;
  };

  return (
    <>
      {HasPerms(user, Permission.EditUser) && (
        <>
          <section className="profile">
            <h3>Profile</h3>
            <form onSubmit={onProfileSubmit}>
              <div className="inputContainer">
                <strong>Name</strong>
                <input
                  className="input"
                  type="text"
                  name="username"
                  placeholder="Required"
                  defaultValue={user.name}
                  minLength={3}
                  maxLength={32}
                />
              </div>
              <button className="button green" type="submit">
                <WithSpinner width="16" height="16" strokeWidth="8" dispatcherRef={setIsUpdatingProfileRef} />
                <strong>Update Profile</strong>
              </button>
            </form>
          </section>

          <section className="password">
            <h3>Password</h3>
            <form onSubmit={onPasswordSubmit}>
              <div className="inputContainer">
                <strong>Current Password</strong>
                <input
                  className="input"
                  type="password"
                  name="currentPassword"
                  placeholder="Required"
                  minLength={6}
                  required
                />
              </div>
              <div className="inputContainer">
                <strong>New Password</strong>
                <input
                  className="input"
                  type="password"
                  name="newPassword"
                  placeholder="Required"
                  minLength={6}
                  required
                />
              </div>
              <button className="button green" type="submit">
                <WithSpinner width="16" height="16" strokeWidth="8" dispatcherRef={setIsUpdatingPasswordRef} />
                <strong>Update Password</strong>
              </button>
            </form>
          </section>
        </>
      )}

      {user.id > 1 && HasPerms(user, Permission.DeleteUser) && (
        <section className="deleteAccount">
          <h3>Account Deletion</h3>
          <form onSubmit={onDeleteAccountSubmit}>
            <div className="alert">
              <AlertTriangle width="16" height="16" strokeWidth="3" />
              <span>
                This operation will permanently delete your user account. It <strong>CAN NOT</strong> be undone.
              </span>
            </div>
            <div className="inputContainer">
              <strong>Password</strong>
              <input className="input" type="password" name="password" placeholder="Required" minLength={6} required />
            </div>
            <button className="button red" type="submit">
              <WithSpinner width="16" height="16" strokeWidth="8" dispatcherRef={setIsDeletingRef} />
              <strong>Confirm Deletion</strong>
            </button>
          </form>
        </section>
      )}
    </>
  );
};

export default User;
