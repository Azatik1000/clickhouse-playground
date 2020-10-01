import React from 'react';
import Axios from 'axios';
import logo from './logo.svg';
import './App.css';

import {useTable} from 'react-table'

class CodeExecutor extends React.Component {
    constructor(props) {
        super(props);
        this.state = {response: null};

        this.handleResponse = this.handleResponse.bind(this);
    }

    handleResponse(response) {
        // console.log(result.data)
        // var historyData = {"result": result};

        // TODO: maybe change data?
        // window.history.pushState("lol", "", response.data.link);

        this.setState({response: JSON.stringify(response.data)});
    }

    render() {
        return (
            <div>
                <CodeForm onResponse={this.handleResponse}/>
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

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleChange(event) {
        this.setState({code: event.target.value});
    }

    async handleSubmit(event) {
        const code = this.state.code;
        // alert(`Submitted ${this.state.value}`);
        // this.setState({value: ''});
        event.preventDefault();

        try {
            const response = await Axios.post("/api/exec", {
                query: code,
                versionID: "v20.9"
            });

            this.props.onResponse(response);
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
