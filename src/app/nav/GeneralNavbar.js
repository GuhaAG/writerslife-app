import React, { Component } from 'react';

const isLoggedIn = window.localStorage.getItem("isLoggedIn");

const handleLogout = () => {
  window.localStorage.removeItem("isLoggedIn");
  window.localStorage.removeItem("loginJwt");

  window.location.replace("/");
}

const redirectToHome = () => {
  window.location.replace("/");
}

class GeneralNavbar extends Component {
  render() {

    return (
      <div>
        <div className="bg-blue-800 m-6 p-6 rounded shadow-lg">

          {(isLoggedIn === "true") &&
            <a className="font-mono text-right self-end rounded bg-blue-800 text-white mr-128"
              href="/Profile">
              My Profile
          </a>
          }

          <button className="ml-56 mr-56" onClick={redirectToHome}>
            <h2 className="text-white text-2xl font-mono">Writerslife</h2>
          </button>

          {(isLoggedIn === "true") &&
            <button className="font-mono text-right self-end rounded bg-blue-800 text-white ml-128"
              onClick={handleLogout}>
              Logout
          </button>

          }

        </div>
      </div>
    )
  }
}

export default GeneralNavbar;
