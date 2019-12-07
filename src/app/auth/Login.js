import React, { Component } from 'react';
import axios from 'axios';
import Swal from 'sweetalert2';

class Login extends Component {

  constructor(props) {
    super(props);

    this.state = {
      username: "",
      email: "",
      password: ""
    };
  }

  componentDidMount() {
    var isLoggedIn = window.localStorage.getItem("isLoggedIn");
    if (isLoggedIn && isLoggedIn === 'true') {
      window.location.replace("/");
    }
  }

  handleChange = event => {
    this.setState({
      [event.target.id]: event.target.value
    });
  }

  handleSubmit = event => {
    event.preventDefault();

    var userId = this.state.username;
    var user;

    if (userId.indexOf("@") === -1) {
      user = {
        username: userId,
        password: this.state.password
      }
    }
    else {
      user = {
        email: userId,
        password: this.state.password
      }
    }

    axios.post('http://localhost:8081/authenticate', user)
      .then((response) => {

        this.setState({
          error: false,
          errorMessage: ''
        });

        window.localStorage.setItem("isLoggedIn", true);
        window.localStorage.setItem("loginJwt", response.data.id_token);

        window.location.replace("/");
      }, (error) => {

        window.localStorage.setItem("isLoggedIn", false);
        window.localStorage.setItem("loginJwt", "");

        let errMessage = error.response.data.message;
        if (error.response.data && error.response.data.errors && error.response.data.errors[0].defaultMessage) {
          errMessage = error.response.data.errors[0].defaultMessage;
        }

        this.setState({
          error: true,
          errorMessage: errMessage
        });

        Swal.fire({
          icon: 'error',
          title: '',
          text: errMessage,
        })

      });
  }

  render() {

    return (

      <div className="bg-grey-lighter flex flex-col">
        <form onSubmit={this.handleSubmit}>
          <div className="container max-w-sm mx-auto flex-1 flex flex-col items-center justify-center px-2">
            <div className="bg-white px-6 py-8 rounded shadow-md text-black w-full font-mono">
              <h1 className="mb-8 text-3xl text-center">Login</h1>

              <input
                required
                type="text"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="username"
                onChange={this.handleChange}
                placeholder="Username/Email" />

              <input
                required
                type="password"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="password"
                onChange={this.handleChange}
                placeholder="Password" />

              {this.state.error && <div>
                <p className="text-red-500">{this.state.errorMessage}</p>
              </div>}

              <button
                type="submit"
                className="w-full text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1"
              >Login</button>

            </div>

            <div className="text-grey-dark mt-6">
              Don't have an account yet ?&nbsp;&nbsp;
                    <a className="no-underline border-b border-blue text-blue" href="../signup/">
                Create one
                    </a>.
                </div>
          </div>
        </form>
      </div>
    );
  }
}

export default Login;
