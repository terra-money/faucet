import React, { Component } from 'react';
import './App.scss';
import SelectNetworks from './components/SelectNetworks';
import Home from './components/Home';

class App extends Component {
  showCurrentYear() {
    return new Date().getFullYear();
  }
  render() {
    return (
      <>
        <header>
          <SelectNetworks className="network_select" />
        </header>
        <Home />
        <footer>
          &copy; 2019-{this.showCurrentYear()} <span>Terra</span>
        </footer>
      </>
    );
  }
}

export default App;
