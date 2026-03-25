--!strict
local DataTypes = require("@lib/data_types")

local TableSchema = {}

export type SingleFieldType = {
	value: DataTypes.FieldValue?,
	label: string,
}

export type FieldGroup = {
	label: string,
	startIdx: number,
	endIdx: number,
}

export type TableSchema = {
	fieldsMap: {[string]: SingleFieldType},
	fieldKeysInOrder: {string},
	label: string,
	groups: {FieldGroup},
}

function TableSchema.makeSingleFieldType(config: {
	value: DataTypes.FieldValue?,
	label: string?,
}?): SingleFieldType
	local c = config or {}
	return {
		value = c.value,
		label = c.label or "",
	}
end

function TableSchema.makeFieldGroup(config: {
	label: string?,
	startIdx: number?,
	endIdx: number?,
}?): FieldGroup
	local c = config or {}
	return {
		label = c.label or "",
		startIdx = c.startIdx or 0,
		endIdx = c.endIdx or 0,
	}
end

function TableSchema.makeTableSchema(config: {
	fieldsMap: {[string]: SingleFieldType}?,
	fieldKeysInOrder: {string}?,
	label: string?,
	groups: {FieldGroup}?,
}?): TableSchema
	local c = config or {}
	return {
		fieldsMap = c.fieldsMap or {},
		fieldKeysInOrder = c.fieldKeysInOrder or {},
		label = c.label or "",
		groups = c.groups or {},
	}
end

function TableSchema.encodeJsonSingleFieldType(sft: SingleFieldType): {[string]: any}
	local result: {[string]: any} = { label = sft.label }
	if sft.value ~= nil then
		result.value = DataTypes.encodeFieldValueByTypeId(sft.value)
	end
	return result
end

function TableSchema.decodeJsonSingleFieldType(json: {[string]: any}): SingleFieldType
	return {
		label = json.label or "",
		value = if json.value ~= nil then DataTypes.decodeFieldValueByTypeId(json.value) else nil,
	}
end

function TableSchema.encodeJsonTableSchema(ts: TableSchema): {[string]: any}
	local fieldsMapEncoded: {[string]: any} = {}
	for key, sft in ts.fieldsMap do
		fieldsMapEncoded[key] = TableSchema.encodeJsonSingleFieldType(sft)
	end
	local groupsEncoded: {any} = {}
	for _, g in ts.groups do
		table.insert(groupsEncoded, { label = g.label, startIdx = g.startIdx, endIdx = g.endIdx })
	end
	return {
		fieldsMap = fieldsMapEncoded,
		fieldKeysInOrder = ts.fieldKeysInOrder,
		label = ts.label,
		groups = groupsEncoded,
	}
end

function TableSchema.decodeJsonTableSchema(json: {[string]: any}): TableSchema
	local fieldsMap: {[string]: SingleFieldType} = {}
	if json.fieldsMap ~= nil then
		for key, sft in json.fieldsMap :: {[string]: any} do
			fieldsMap[key] = TableSchema.decodeJsonSingleFieldType(sft)
		end
	end
	local groups: {FieldGroup} = {}
	if json.groups ~= nil then
		for _, g in json.groups :: {{[string]: any}} do
			table.insert(groups, {
			label = g.label or "",
				startIdx = g.startIdx or 0,
				endIdx = g.endIdx or 0,
			})
		end
	end
	return {
		fieldsMap = fieldsMap,
		fieldKeysInOrder = json.fieldKeysInOrder or {},
		label = json.label or "",
		groups = groups,
	}
end

return TableSchema
