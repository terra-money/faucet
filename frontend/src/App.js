import React, { Component } from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import Route from './route';
import './App.scss';
import SelectNetworks from './components/SelectNetworks';

class App extends Component {
  showCurrentYear() {
    return new Date().getFullYear();
  }
  render() {
    return (
      <Router>
        <header>
          <SelectNetworks className="network_select" />
        </header>
        <Route />
        <footer>
          &copy; {this.showCurrentYear()} <span>Terra</span>
        </footer>
      </Router>
    );
  }
}

export default App;
