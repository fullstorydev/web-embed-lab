/**
DOM shape test probe
*/
class DOMShapeProbe {
	/**
	@return {Object} data collected when the target embed script *is not* loaded
	@return {Object.success} true if the data collection was successful
	@return {Object.width} DOM width
	@return {Object.depth} DOM depth
	*/
	async gatherBaselineData(){
		console.log('Baseline DOM shape')
		const [width, depth] = this._findWidthAndDepth()
		return {
			success: true,
			width: width,
			depth: depth
		}
	}

	/**
	@return {object} the results of the probe
	*/
	async probe(basis, baseline){
		console.log('Probing DOM shape', baseline)
		if(!basis) return { passed: true }

		const [width, depth] = this._findWidthAndDepth()
		const results = {
			passed: true,
			depth: depth,
			width: width
		}
		if(typeof basis.depth === 'number' && basis.depth !== results.depth){
			results.passed = false
		} else if (Array.isArray(basis.depth)) {
			if(basis.depth.length !== 2 || basis.depth[0] > basis.depth[1]){
				console.error('Invalid depth range: ' + basis.depth)
				results.passed = false
				return results
			}
			if(basis.depth[0] > results.depth || basis.depth[1] < results.depth){
				console.error('Depth range failed for ' + results.depth + ': ' + basis.depth)
				results.passed = false
			}
		}

		if(typeof basis.width === 'number' && basis.width !== results.width){
			results.passed = false
		} else if(Array.isArray(basis.width)){
			if(basis.width.length !== 2 || basis.width[0] > basis.width[1]){
				console.error('Invalid width range: ' + basis.width)
				results.passed = false
				return results
			}
			if(basis.width[0] > results.width || basis.width[1] < results.width){
				console.error('Width range failed for ' + results.width + ': ' + basis.width)
				results.passed = false
			}
		}
		return results
	}

	// returns [width, depth]
	_findWidthAndDepth(){
		const shape = this._findShape(document.body)
		let width = 0;
		for(let i=0; i < shape.rows.length; i++){
			width = Math.max(width, shape.rows[i].length)
		}
		return [width, shape.rows.length]
	}

	_findShape(element, depth=0, results={ rows: [] }){
		if(!results.rows[depth]) results.rows[depth] = []
		results.rows[depth].push(element.children.length)
		for(let i=0; i < element.children.length; i++){
			this._findShape(element.children[i], depth + 1, results)
		}
		return results
	}
}

window.__welProbes['dom-shape'] = new DOMShapeProbe()

