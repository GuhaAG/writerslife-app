import React, { Component } from 'react';
import GeneralNavbar from './nav/GeneralNavbar';
import Main from './Main';

class App extends Component {
  render() {
    return (
      <div>
        <GeneralNavbar />
        <Main />
      </div>
    );
  }
}

export default App;
