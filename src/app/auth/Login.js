import React, { Component } from 'react';

class Login extends Component {

  constructor(props) {
    super(props);

    this.state = {
      username: "",
      email: "",
      password: ""     
    };
  }

  handleChange = event => {
    this.setState({
      [event.target.id]: event.target.value
    });
  }

  handleSubmit = event => {
    event.preventDefault();    
  }

  render() {

    return (
      
      <div className="bg-grey-lighter flex flex-col">
        <form onSubmit={this.handleSubmit}>
          <div className="container max-w-sm mx-auto flex-1 flex flex-col items-center justify-center px-2">
            <div className="bg-white px-6 py-8 rounded shadow-md text-black w-full font-mono">
              <h1 className="mb-8 text-3xl text-center">Login</h1>              
              
              <input
                type="text"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="username"
                onChange={this.handleChange}
                placeholder="Username/Email" />   

              <input
                type="password"
                className="block border border-grey-light w-full p-3 rounded mb-4"
                id="password"
                onChange={this.handleChange}
                placeholder="Password" />

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
