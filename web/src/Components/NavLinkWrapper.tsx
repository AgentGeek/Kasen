import React from "react";
import { NavLink, NavLinkProps } from "react-router-dom";

const NavLinkWrapper = (props: NavLinkProps) => (
  <NavLink isActive={(_, { pathname, search }) => `${pathname}${search}` === props.to} {...props} />
);

export default NavLinkWrapper;
