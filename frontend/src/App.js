import React from 'react';
import Axios from 'axios';
import logo from './logo.svg';
import './App.css';

import {useTable} from 'react-table'

import {
    BrowserRouter as Router,
    Switch,
    Route,
    Link
} from "react-router-dom";

class CodeExecutor extends React.Component {
    constructor(props) {
        super(props);
        this.state = {response: null};

        // this.handleResponse = this.handleResponse.bind(this);
        this.handleExecute = this.handleExecute.bind(this);
    }

    handleResponse(response) {
        // console.log(result.data)
        // var historyData = {"result": result};

        // TODO: maybe change data?
        // window.history.pushState("lol", "", response.data.link);
    }

    async handleExecute(code, versionID) {
        try {
            const response = await Axios.post("/api/exec", {
                query: code,
                versionID: versionID
            });

            this.setState({response: JSON.stringify(response.data)});
        } catch (e) {
            console.log(e);
        }
    }

    render() {
        return (
            <div>
                <CodeForm onSubmit={this.handleExecute}/>
                <ResultView response={this.state.response}/>
            </div>
        )
    }
}

function Table(props) {
    // console.log(props);
    const {data} = props;
    console.log(data.data);

    const meta = data.meta

    const columns = React.useMemo(() => {
            return meta.map(kek => {
                return {
                    Header: kek.name,
                    accessor: kek.name,
                }
                // console.log(kek);
            });
        },
        [meta]
    )

    const tableInstance = useTable({columns, data: data.data});

    const {
        getTableProps,
        getTableBodyProps,
        headerGroups,
        rows,
        prepareRow,
    } = tableInstance

    return (
        <table {...getTableProps()}>
            <thead>
            {
                headerGroups.map(headerGroup => (
                    <tr {...headerGroup.getHeaderGroupProps()}>
                        {
                            headerGroup.headers.map(column => (
                                <th {...column.getHeaderProps()}>
                                    {
                                        column.render('Header')
                                    }
                                </th>
                            ))
                        }
                    </tr>
                ))
            }
            </thead>
            <tbody {...getTableBodyProps()}>
            {
                rows.map(row => {
                    prepareRow(row)

                    return (
                        <tr {...row.getRowProps()}>
                            {
                                row.cells.map(cell => {
                                    return (
                                        <td
                                            {...cell.getCellProps()}
                                            style={{
                                                padding: '10px',
                                                border: 'solid 1px gray',
                                                background: 'papayawhip',
                                            }}
                                        >
                                            {cell.render('Cell')}
                                        </td>
                                    )
                                })
                            }
                        </tr>
                    )
                })
            }
            </tbody>
        </table>
    )
}

class ResultView extends React.Component {
    constructor(props) {
        super(props);
    }

    render() {
        // TODO: maybe change init value from null to something else
        if (this.props.response === null) {
            return (
                <p>No results yet</p>
            )
        }

        const results = JSON.parse(this.props.response).result;

        return (
            <div>
                <p>
                    {this.props.response}
                </p>
                <Table data={results[0]}/>
            </div>
        )
    }
}

class CodeForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {code: ''};

        this.handleCodeChange = this.handleCodeChange.bind(this);
        this.handleVersionChange = this.handleVersionChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleCodeChange(event) {
        this.setState({code: event.target.value});
    }

    handleVersionChange(event) {
        this.setState({versionID: event.target.value})
    }

    async handleSubmit(event) {
        const code = this.state.code;
        const versionID = this.state.versionID
        // alert(`Submitted ${this.state.value}`);
        // this.setState({value: ''});
        event.preventDefault();

        await this.props.onSubmit(code, versionID)
    }

    render() {
        return (
            <form onSubmit={this.handleSubmit}>
                <label>
                    <textarea cols="40" rows="5" value={this.state.code} onChange={this.handleCodeChange}/>
                </label>

                <input type="radio" name="version" id="v20.9" onChange={this.handleVersionChange} value="v20.9" />
                <label htmlFor="v20.9">v20.9</label>

                <input type="radio" name="version" id="v20.8" onChange={this.handleVersionChange} value="v20.8" />
                <label htmlFor="v20.8">v20.8</label>

                <input type="submit" value="Submit"/>
            </form>
        )
    }
}



function App() {
    return (
        <Router>
            <div>
                <Switch>
                    <Route path="/runs/:id">
                        <div className="App">
                            <h1>ClickHouse Explorer</h1>
                            <CodeExecutor/>
                        </div>
                    </Route>
                    <Route path="/">
                        <div className="App">
                            <h1>ClickHouse Explorer</h1>
                            <CodeExecutor/>
                        </div>
                    </Route>
                </Switch>
            </div>
        </Router>
    );
}

export default App;
