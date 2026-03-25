--!strict
local DataTypes = {}

export type StringType = {
	typeId: string,
	value: string,
	regex: string?,
	formatPatterns: {string},
}

export type NumberType = {
	typeId: string,
	value: number,
	decimalPlaces: number,
	minValue: number?,
	maxValue: number?,
}

export type BooleanType = {
	typeId: string,
	value: boolean,
}

export type EnumType = {
	typeId: string,
	value: string,
	options: {string},
}

export type MultipleCheckboxType = {
	typeId: string,
	selectedKeys: {string},
	allOptionKeysInOrder: {string},
	allOptionLabelsInOrder: {string},
}

export type RadioGroupType = {
	typeId: string,
	selectedKey: string,
	allOptionKeysInOrder: {string},
	allOptionLabelsInOrder: {string},
}

function DataTypes.makeStringType(config: {
	typeId: string?,
	value: string?,
	regex: string?,
	formatPatterns: {string}?,
}?): StringType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		value = c.value or "",
		regex = c.regex,
		formatPatterns = c.formatPatterns or {},
	}
end

function DataTypes.makeNumberType(config: {
	typeId: string?,
	value: number?,
	decimalPlaces: number?,
	minValue: number?,
	maxValue: number?,
}?): NumberType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		value = c.value or 0,
		decimalPlaces = c.decimalPlaces or 0,
		minValue = c.minValue,
		maxValue = c.maxValue,
	}
end

function DataTypes.makeBooleanType(config: {
	typeId: string?,
	value: boolean?,
}?): BooleanType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		value = if c.value ~= nil then c.value else false,
	}
end

function DataTypes.makeEnumType(config: {
	typeId: string?,
	value: string?,
	options: {string}?,
}?): EnumType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		value = c.value or "",
		options = c.options or {},
	}
end

function DataTypes.makeMultipleCheckboxType(config: {
	typeId: string?,
	selectedKeys: {string}?,
	allOptionKeysInOrder: {string}?,
	allOptionLabelsInOrder: {string}?,
}?): MultipleCheckboxType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		selectedKeys = c.selectedKeys or {},
		allOptionKeysInOrder = c.allOptionKeysInOrder or {},
		allOptionLabelsInOrder = c.allOptionLabelsInOrder or {},
	}
end

function DataTypes.makeRadioGroupType(config: {
	typeId: string?,
	selectedKey: string?,
	allOptionKeysInOrder: {string}?,
	allOptionLabelsInOrder: {string}?,
}?): RadioGroupType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		selectedKey = c.selectedKey or "",
		allOptionKeysInOrder = c.allOptionKeysInOrder or {},
		allOptionLabelsInOrder = c.allOptionLabelsInOrder or {},
	}
end

function DataTypes.encodeJsonStringType(s: StringType): {[string]: any}
	local result: {[string]: any} = {
		typeId = s.typeId,
		value = s.value,
		formatPatterns = s.formatPatterns,
	}
	if s.regex ~= nil then
		result.regex = s.regex
	end
	return result
end

function DataTypes.decodeJsonStringType(json: {[string]: any}): StringType
	return {
		typeId = json.typeId or "",
		value = json.value or "",
		regex = json.regex,
		formatPatterns = json.formatPatterns or {},
	}
end

function DataTypes.encodeJsonNumberType(n: NumberType): {[string]: any}
	local result: {[string]: any} = {
		typeId = n.typeId,
		value = n.value,
		decimalPlaces = n.decimalPlaces,
	}
	if n.minValue ~= nil then
		result.minValue = n.minValue
	end
	if n.maxValue ~= nil then
		result.maxValue = n.maxValue
	end
	return result
end

function DataTypes.decodeJsonNumberType(json: {[string]: any}): NumberType
	return {
		typeId = json.typeId or "",
		value = json.value or 0,
		decimalPlaces = json.decimalPlaces or 0,
		minValue = json.minValue,
		maxValue = json.maxValue,
	}
end

function DataTypes.encodeJsonBooleanType(b: BooleanType): {[string]: any}
	return {
		typeId = b.typeId,
		value = b.value,
	}
end

function DataTypes.decodeJsonBooleanType(json: {[string]: any}): BooleanType
	return {
		typeId = json.typeId or "",
		value = if json.value ~= nil then json.value else false,
	}
end

function DataTypes.encodeJsonEnumType(e: EnumType): {[string]: any}
	return {
		typeId = e.typeId,
		value = e.value,
		options = e.options,
	}
end

function DataTypes.decodeJsonEnumType(json: {[string]: any}): EnumType
	return {
		typeId = json.typeId or "",
		value = json.value or "",
		options = json.options or {},
	}
end

function DataTypes.encodeJsonMultipleCheckboxType(m: MultipleCheckboxType): {[string]: any}
	return {
		typeId = m.typeId,
		selectedKeys = m.selectedKeys,
		allOptionKeysInOrder = m.allOptionKeysInOrder,
		allOptionLabelsInOrder = m.allOptionLabelsInOrder,
	}
end

function DataTypes.decodeJsonMultipleCheckboxType(json: {[string]: any}): MultipleCheckboxType
	return {
		typeId = json.typeId or "",
		selectedKeys = json.selectedKeys or {},
		allOptionKeysInOrder = json.allOptionKeysInOrder or {},
		allOptionLabelsInOrder = json.allOptionLabelsInOrder or {},
	}
end

function DataTypes.encodeJsonRadioGroupType(r: RadioGroupType): {[string]: any}
	return {
		typeId = r.typeId,
		selectedKey = r.selectedKey,
		allOptionKeysInOrder = r.allOptionKeysInOrder,
		allOptionLabelsInOrder = r.allOptionLabelsInOrder,
	}
end

function DataTypes.decodeJsonRadioGroupType(json: {[string]: any}): RadioGroupType
	return {
		typeId = json.typeId or "",
		selectedKey = json.selectedKey or "",
		allOptionKeysInOrder = json.allOptionKeysInOrder or {},
		allOptionLabelsInOrder = json.allOptionLabelsInOrder or {},
	}
end

return DataTypes
