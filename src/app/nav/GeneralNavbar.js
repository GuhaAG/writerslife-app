import React, { Component } from 'react';
import { Container, Navbar, Nav, NavDropdown } from 'react-bootstrap';
import 'bootstrap/dist/css/bootstrap.min.css';

const isLoggedIn = window.localStorage.getItem("isLoggedIn");
const username = window.localStorage.getItem("username");

const handleLogout = () => {
  window.localStorage.removeItem("isLoggedIn");
  window.localStorage.removeItem("loginJwt");
  window.localStorage.removeItem("username");

  window.location.replace("/");
}

class GeneralNavbar extends Component {
  render() {

    return (
      <div>
        <Container>
          <Navbar className="bg-blue-800 text-center text-white font-mono my-6 p-6 rounded shadow-lg" variant="dark">
            <Navbar.Brand href="/">Writerslife</Navbar.Brand>
            <Navbar.Toggle aria-controls="responsive-navbar-nav" />
            <Navbar.Collapse id="responsive-navbar-nav">
              <Nav className="mr-auto">
                {(isLoggedIn === "true") &&
                  <NavDropdown title={username} id="collasible-nav-dropdown">
                    <NavDropdown.Item href="/Profile">Profile</NavDropdown.Item>
                    <NavDropdown.Item href="/Settings">Settings</NavDropdown.Item>
                  </NavDropdown>
                }
              </Nav>
              <Nav>
                <Nav.Link onClick={handleLogout}>Logout</Nav.Link>
              </Nav>
            </Navbar.Collapse>
          </Navbar>
        </Container>
      </div>
    )
  }
}

export default GeneralNavbar;
