import React from 'react';
// import cx from "classnames";
import ReCAPTCHA from 'react-google-recaptcha';
import axios from 'axios';
import { ToastContainer, toast } from 'react-toastify';
import { Formik, Form, Field } from 'formik';
import * as Yup from 'yup';
import b32 from '../../lib/b32';
import { networks } from '../../config';

import 'react-toastify/dist/ReactToastify.css';

import '../../App.scss';
import NetworkContext from '../../contexts/NetworkContext';

const bech32Validate = (param) => {
  try {
    b32.decode(param);
  } catch (error) {
    return error.message;
  }
};

const sendSchema = Yup.object().shape({
  address: Yup.string().required('Required'),
  denom: Yup.string().required('Required'),
});

const DENUMS_TO_TOKEN = {
  uluna: 'Luna',
  uusd: 'UST',
  ueur: 'EUT',
  usdr: 'SDT',
  ukrw: 'KRT',
};

const REQUEST_LIMIT_SECS = 30;

class HomeComponent extends React.Component {
  static contextType = NetworkContext;
  recaptchaRef = React.createRef();

  constructor(props) {
    super(props);
    this.state = {
      sending: false,
      verified: false,
      response: '',
    };
  }

  handleCaptcha = (response) => {
    this.setState({
      response,
      verified: true,
    });
  };

  handleSubmit = (values, { resetForm }) => {
    const network = this.context.network;
    const item = networks.filter((n) => n.key === network)[0];
    // same shape as initial values
    this.setState({
      sending: true,
      verified: false,
    });

    this.recaptchaRef.current.reset();

    setTimeout(() => {
      this.setState({ sending: false });
    }, REQUEST_LIMIT_SECS * 1000);

    axios
      .post('https://faucet.terra.dev/claim', {
        chain_id: network,
        lcd_url: item.lcd,
        address: values.address,
        denom: values.denom,
        response: this.state.response,
      })
      .then((response) => {
        const { amount } = response.data;

        toast.success(
          `Successfully Sent ${amount / 1000000} ${
            DENUMS_TO_TOKEN[values.denom]
          } to ${values.address}`
        );

        resetForm();
      })
      .catch((err) => {
        let errText;

        switch (err.response.status) {
          case 400:
            errText = 'Invalid request';
            break;
          case 403:
            errText = 'Too many requests';
            break;
          case 404:
            errText = 'Cannot connect to server';
            break;
          case 502:
          case 503:
            errText = 'Faucet service temporary unavailable';
            break;
          default:
            errText = err.response.data || err.message;
            break;
        }

        toast.error(`An error occurred: ${errText}`);
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
            <ReCAPTCHA
              ref={this.recaptchaRef}
              sitekey="6Ld4w4cUAAAAAJceMYGpOTpjiJtMS_xvzOg643ix"
              onChange={this.handleCaptcha}
            />
          </div>
          <Formik
            initialValues={{
              address: '',
              denom: '',
            }}
            validationSchema={sendSchema}
            onSubmit={this.handleSubmit}
          >
            {({ errors, touched }) => (
              <Form className="inputContainer">
                <div className="input">
                  <Field
                    name="address"
                    placeholder="Testnet address"
                    validate={bech32Validate}
                  />
                  {errors.address && touched.address ? (
                    <div className="fieldError">{errors.address}</div>
                  ) : null}
                </div>
                <div className="select">
                  <Field component="select" name="denom">
                    <option value="" default>
                      Select denom to receive...
                    </option>
                    <option value="uluna">{DENUMS_TO_TOKEN['uluna']}</option>
                    <option value="ukrw">{DENUMS_TO_TOKEN['ukrw']}</option>
                    <option value="uusd">{DENUMS_TO_TOKEN['uusd']}</option>
                    <option value="usdr">{DENUMS_TO_TOKEN['usdr']}</option>
                    <option value="umnt">{DENUMS_TO_TOKEN['umnt']}</option>
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
                    disabled={!this.state.verified || this.state.sending}
                    type="submit"
                  >
                    <i aria-hidden="true" className="material-icons">
                      send
                    </i>
                    <span>
                      {this.state.sending
                        ? 'Waiting for next tap'
                        : 'Send me tokens'}
                    </span>
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
                  href="https://jigu.terra.money"
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
