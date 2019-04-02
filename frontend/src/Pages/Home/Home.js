import React from "react";
// import cx from "classnames";
import Reaptcha from "reaptcha";
import axios from "axios";
import { ToastContainer, toast } from "react-toastify";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import b32 from "../../scripts/b32";

import "react-toastify/dist/ReactToastify.css";

import "../../App.scss";

const bech32Validate = param => {
  try {
    b32.decode(param);
  } catch (error) {
    return error.message;
  }
};

const sendSchema = Yup.object().shape({
  address: Yup.string().required("Required"),
  denom: Yup.string().required("Required")
});

class HomeComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      sending: false,
      verified: false
    };
  }

  onVerify = () => {
    this.setState({
      verified: true
    });
  };

  render() {
    return (
      <div>
        <ToastContainer
          position="top-right"
          autoClose={5000}
          hideProgressBar
          newestOnTop
          closeOnClick
          rtl={false}
          pauseOnVisibilityChange
          pauseOnHover
        />
        <section>
          <h2>Terra Testnet Faucet</h2>
          <article>
            Hello intrepid spaceperson! Use this faucet to get tokens for the
            latest Terra testnet. Please don't abuse this service—the number of
            available tokens is limited.
          </article>
          <div className="recaptcha">
            <Reaptcha
              sitekey="6Ld4w4cUAAAAAJceMYGpOTpjiJtMS_xvzOg643ix"
              onVerify={this.onVerify}
            />
          </div>
          <Formik
            initialValues={{
              address: "",
              denom: ""
            }}
            validationSchema={sendSchema}
            onSubmit={(values, { resetForm }) => {
              // same shape as initial values
              this.setState({ sending: true });
              axios
                .post("/claim", {
                  address: values.address,
                  denom: values.denom
                })
                .then(() => {
                  this.setState({ sending: false });
                  toast.success(
                    `Successfully Sent, Sent tokens to ${this.fields.address}`
                  );

                  resetForm();
                })
                .catch(err => {
                  this.setState({ sending: false });
                  toast.error(
                    `Error Sending, An error occurred while trying to send: "${
                      err.message
                    }"`
                  );
                });
            }}
          >
            {({ errors, touched }) => (
              <Form className="inputContainer">
                <div className="input">
                  <Field name="address" validate={bech32Validate} />
                  {errors.address && touched.address ? (
                    <div className="fieldError">{errors.address}</div>
                  ) : null}
                </div>
                <div className="select">
                  <Field className="select" component="select" name="denom">
                    <option
                      value=""
                      disabled="disabled"
                      selected="selected"
                      hidden="hidden"
                    >
                      Select denom to receive...
                    </option>
                    <option value="mluna">Luna</option>
                    <option value="mkrw">KRW</option>
                    <option value="musd">USD</option>
                    <option value="msdr">SDR</option>
                    <option value="mgbp">GBP</option>
                    <option value="meur">EUR</option>
                    <option value="mjpy">JPY</option>
                    <option value="mcny">CNY</option>
                  </Field>
                  {errors.denom && touched.denom ? (
                    <div className="fieldError selectFieldError">
                      {errors.denom}
                    </div>
                  ) : null}
                  <div className="selectAddon">
                    <i className="material-icons">arrow_drop_down</i>
                  </div>
                </div>

                <div className="buttonContainer">
                  <button
                    disabled={!this.state.verified}
                    type="submit"
                  >
                    <i aria-hidden="true" className="material-icons">
                      send
                    </i>
                    <span>Send me tokens</span>
                  </button>
                </div>
              </Form>
            )}
          </Formik>
        </section>
        <section>
          <h2>Don't you have a testnet address?</h2>
          <article>
            There's two ways to get one. The first is by using Station, the
            crypto wallet for Terra. If you know command-line-fu, you can also
            generate an address with the Terra SDK.
          </article>
          <div className="buttonContainer">
            <button className="light">
              <i aria-hidden="true" className="material-icons">
                supervisor_account
              </i>
              <span>
                <a
                  href="https://github.com/terra-project/core/blob/develop/docs/guide/join-network.md"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Join the latest testnet
                </a>
              </span>
            </button>
          </div>
        </section>
      </div>
    );
  }
}

export default HomeComponent;
