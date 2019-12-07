import React from 'react';
import { Redirect, Route } from 'react-router-dom';

// This is used to determine if a user is authenticated and
// if they are allowed to visit the page they navigated to.

// If they are: they proceed to the page
// If not: they are redirected to the login page.

const PrivateRoute = ({ component: Component, ...rest }) => {

  const token = window.localStorage.getItem("loginJwt");
  const isLoggedIn = window.localStorage.getItem("isLoggedIn");

  if (token == null || token.length === 0 || !isLoggedIn) {
    window.localStorage.setItem("isLoggedIn", false);
  }

  return (
    <Route
      {...rest}
      render={props =>
        (isLoggedIn === 'true') ? (
          <Component {...props} />
        ) : (
            <Redirect to={{ pathname: '/login', state: { from: props.location } }} />
          )
      }
    />
  )
}

export default PrivateRoute
