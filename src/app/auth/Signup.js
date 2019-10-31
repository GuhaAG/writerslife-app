import React, { Component } from 'react';

class Signup extends Component {

  constructor(props) {
    super(props);

    this.state = {
      username: "",
      firstname: "",
      lastname: "",
      email: "",
      password: "",
      confirm_password: ""
    };
  }

  handleChange = event => {
    this.setState({
      [event.target.id]: event.target.value
    });
  }

  handleSubmit = event => {
    event.preventDefault();

    /* var user = {
      username: this.state.username,
      email: this.state.email,
      firstname: this.state.firstname,
      lastname: this.state.lastname,
      password: this.state.password,
      confirm_password: this.state.confirm_password
    }

    console.log(user); */
  }

  render() {

    return (
      
      <div className="bg-grey-lighter flex flex-col">
        <form onSubmit={this.handleSubmit}>
          <div className="container max-w-sm mx-auto flex-1 flex flex-col items-center justify-center px-2">
            <div className="bg-white px-6 py-8 rounded shadow-md text-black w-full font-mono">
              <h1 className="mb-8 text-3xl text-center">Sign up</h1>              
              <input
                type="text"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="firstname"
                onChange={this.handleChange}
                placeholder="First Name" />

              <input
                type="text"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="lastname"
                onChange={this.handleChange}
                placeholder="Last Name" />  

              <input
                type="text"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="email"
                onChange={this.handleChange}
                placeholder="Email" />

              <input
                type="text"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="username"
                onChange={this.handleChange}
                placeholder="Username" />   

              <input
                type="password"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="password"
                onChange={this.handleChange}
                placeholder="Password" />

              <input
                type="password"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="confirm_password"
                onChange={this.handleChange}
                placeholder="Confirm Password" />

              <button
                type="submit"
                className="w-full text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1"
              >Create Account</button>

            </div>

            <div className="text-grey-dark mt-6">
              Already have an account?&nbsp;&nbsp;
                    <a className="no-underline border-b border-blue text-blue" href="../login/">
                Log in
                    </a>.
                </div>
          </div>
        </form>    
      </div>
    );
  }
}

export default Signup;