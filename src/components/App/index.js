import React from 'react';
import CoinCard from '../CoinCard';

import './index.scss'

class App extends React.Component {
	state = {
		coins: null,
		error: null,
	}

	componentDidMount(){
		const endpoints = [
			`${process.env.REACT_APP_BACKEND}/eth`,
			`${process.env.REACT_APP_BACKEND}/btc`,
		]

		const symbols = [
			"eth",
			"btc"
		]

		Promise.all(endpoints.map(e => fetch(e)))
		.then(responses => Promise.all(responses.map(r => r.json())))
		.then(data => {
			const coins = {}
			symbols.forEach((symbol, i) => {
				coins[symbol] = {exchanges: data[i]}
				coins[symbol].loaded = true
			})
			this.setState({coins: coins})
		})
		.catch(error => this.setState({error: "Cannot make initial fetch"}))
}

	render(){
		return <div className="app">
			<div className="container">
			<div className="title">
				<h1>CoinPicker</h1>
			</div>

			{this.state.error ?
				<div className="error">
					<p>Something went wrong. <span onClick={
						()=> window.location.reload()
					}>Retry</span></p>
				</div>
			: null}

			<div className="coins">
			{
				!this.state.coins ? null :
				Object.keys(this.state.coins).map(coin => <CoinCard
					name={coin}
					icon={`/${coin}.svg`}
					exchanges={this.state.coins[coin].exchanges}
					loaded={this.state.coins[coin].loaded}
					refreshHandler={()=>{
						const coins = this.state.coins
						coins[coin].loaded = false
						this.setState({coins: coins})

						fetch(`${process.env.REACT_APP_BACKEND}/${coin}`)
						.then(response => response.json())
						.then(data => {
							const coins = this.state.coins
							coins[coin].exchanges=data
							coins[coin].loaded = true 
							this.setState({coins: coins})
						})
						.catch(error => this.setState({error: `Cannot fetch ${process.env.REACT_APP_BACKEND}/${coin}`}))
					}}
				/>)
			}
			</div>
			</div>
		</div>
	}
}

export default App