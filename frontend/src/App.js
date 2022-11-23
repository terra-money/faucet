import React, { Component } from 'react';
import './App.scss';
import Header from './components/Header/Header';
import Home from './components/Home';

class App extends Component {
  showCurrentYear() {
    return new Date().getFullYear();
  }
  render() {
    return (
      <>
        <Header />
        <Home />
        <footer>
          &copy; 2019-{this.showCurrentYear()} <span>Terra</span>
        </footer>
      </>
    );
  }
}

export default App;
