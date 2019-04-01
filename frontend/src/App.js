import React, { Component } from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import Route from './route';
import './App.scss';

class App extends Component {
  showCurrentYear() {
    return new Date().getFullYear();
  }
  render() {
    return (
      <Router>
        <header>
          testnet: <span>soju-0005</span>
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
