import React from 'react';
import Axios from 'axios';
import logo from './logo.svg';
import './App.css';

class CodeExecutor extends React.Component {
    constructor(props) {
        super(props);
        this.state = {result: ''};

        this.handleResult = this.handleResult.bind(this);
    }

    handleResult(result) {
        this.setState({result: result.data.result});
    }

    render() {
        return (
            <div>
                <CodeForm onResult={this.handleResult}/>
                <ResultView result={this.state.result}/>
            </div>
        )
    }
}

class ResultView extends React.Component {
    constructor(props) {
        super(props);
    }

    render() {
        return (
            <p>
                {this.props.result}
            </p>
        )
    }
}

class CodeForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {code: ''};

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleChange(event) {
        this.setState({code: event.target.value});
    }

    async handleSubmit(event) {
        const code = this.state.code;
        // alert(`Submitted ${this.state.value}`);
        this.setState({value: ''});
        event.preventDefault();

        try {
            const response = await Axios.post("http://localhost:3999/exec", {
                Query: code
            });

            this.props.onResult(response);
        } catch (e) {
            console.log(e);
        }
    }

    render() {
        return (
            <form onSubmit={this.handleSubmit}>
                <label>
                    Code:
                    <textarea cols="40" rows="5" value={this.state.code} onChange={this.handleChange}/>
                </label>
                <input type="submit" value="Submit"/>
            </form>
        )
    }
}

function App() {
    return (
        <div className="App">
            <h1>ClickHouse Playground</h1>
            <CodeExecutor/>
        </div>
    );
}

export default App;
