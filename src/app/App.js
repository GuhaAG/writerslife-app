import React, { Component } from 'react';
import GeneralNavbar from './nav/GeneralNavbar';
import Main from './Main';

class App extends Component {
  render() {
    return (
      <div className="text-center">
        <GeneralNavbar />
        <Main />
      </div>
    );
  }
}

export default App;
