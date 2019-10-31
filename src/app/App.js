import React, { Component } from 'react';
import Main from './Main';

class App extends Component {
  render() {
    return (
      <div className="text-center">
        <div className="bg-blue-800 m-6 p-6 rounded shadow-lg">
          <h2 className="text-white text-2xl font-mono">Writerslife</h2>
        </div>       
        <Main />        
      </div>
    );
  }
}

export default App;
