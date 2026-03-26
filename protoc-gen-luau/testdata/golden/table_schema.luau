--!strict
local DataTypes = require("@lib/data_types")

local TableSchema = {}

export type FieldGroup = {
	label: string,
	startIdx: number,
	endIdx: number,
}

export type SingleFieldType = {
	stringType: DataTypes.StringType?,
	numberType: DataTypes.NumberType?,
	booleanType: DataTypes.BooleanType?,
	enumType: DataTypes.EnumType?,
	multipleCheckboxType: DataTypes.MultipleCheckboxType?,
	radioGroupType: DataTypes.RadioGroupType?,
	phoneFaxType: DataTypes.PhoneFaxType?,
	dateTimeType: DataTypes.DateTimeType?,
	moneyType: DataTypes.MoneyType?,
	addressType: DataTypes.AddressType?,
	individualNameType: DataTypes.IndividualNameType?,
	countryType: DataTypes.CountryType?,
	baseContactType: DataTypes.BaseContactType?,
	submissionContactType: DataTypes.SubmissionContactType?,
	signatoryType: DataTypes.SignatoryType?,
	bankInfoType: DataTypes.BankInfoType?,
	bankAccountInfoType: DataTypes.BankAccountInfoType?,
	wireInstructionsType: DataTypes.WireInstructionsType?,
	brokerageFirmType: DataTypes.BrokerageFirmType?,
	brokerageAccountType: DataTypes.BrokerageAccountType?,
	serviceContactPointType: DataTypes.ServiceContactPointType?,
	customCompoundType: DataTypes.CustomCompoundType?,
	label: string,
}

export type TableSchema = {
	fieldsMap: {[string]: SingleFieldType},
	fieldKeysInOrder: {string},
	label: string,
	groups: {FieldGroup},
}

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

function TableSchema.makeSingleFieldType(config: {
	stringType: DataTypes.StringType?,
	numberType: DataTypes.NumberType?,
	booleanType: DataTypes.BooleanType?,
	enumType: DataTypes.EnumType?,
	multipleCheckboxType: DataTypes.MultipleCheckboxType?,
	radioGroupType: DataTypes.RadioGroupType?,
	phoneFaxType: DataTypes.PhoneFaxType?,
	dateTimeType: DataTypes.DateTimeType?,
	moneyType: DataTypes.MoneyType?,
	addressType: DataTypes.AddressType?,
	individualNameType: DataTypes.IndividualNameType?,
	countryType: DataTypes.CountryType?,
	baseContactType: DataTypes.BaseContactType?,
	submissionContactType: DataTypes.SubmissionContactType?,
	signatoryType: DataTypes.SignatoryType?,
	bankInfoType: DataTypes.BankInfoType?,
	bankAccountInfoType: DataTypes.BankAccountInfoType?,
	wireInstructionsType: DataTypes.WireInstructionsType?,
	brokerageFirmType: DataTypes.BrokerageFirmType?,
	brokerageAccountType: DataTypes.BrokerageAccountType?,
	serviceContactPointType: DataTypes.ServiceContactPointType?,
	customCompoundType: DataTypes.CustomCompoundType?,
	label: string?,
}?): SingleFieldType
	local c = config or {}
	return {
		stringType = c.stringType,
		numberType = c.numberType,
		booleanType = c.booleanType,
		enumType = c.enumType,
		multipleCheckboxType = c.multipleCheckboxType,
		radioGroupType = c.radioGroupType,
		phoneFaxType = c.phoneFaxType,
		dateTimeType = c.dateTimeType,
		moneyType = c.moneyType,
		addressType = c.addressType,
		individualNameType = c.individualNameType,
		countryType = c.countryType,
		baseContactType = c.baseContactType,
		submissionContactType = c.submissionContactType,
		signatoryType = c.signatoryType,
		bankInfoType = c.bankInfoType,
		bankAccountInfoType = c.bankAccountInfoType,
		wireInstructionsType = c.wireInstructionsType,
		brokerageFirmType = c.brokerageFirmType,
		brokerageAccountType = c.brokerageAccountType,
		serviceContactPointType = c.serviceContactPointType,
		customCompoundType = c.customCompoundType,
		label = c.label or "",
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

local function decodeOptionalAddressType(json: any): DataTypes.AddressType?
	if json == nil then return nil end
	return DataTypes.decodeJsonAddressType(json)
end
local function encodeOptionalAddressType(v: DataTypes.AddressType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonAddressType(v)
end
local function decodeOptionalBankAccountInfoType(json: any): DataTypes.BankAccountInfoType?
	if json == nil then return nil end
	return DataTypes.decodeJsonBankAccountInfoType(json)
end
local function encodeOptionalBankAccountInfoType(v: DataTypes.BankAccountInfoType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonBankAccountInfoType(v)
end
local function decodeOptionalBankInfoType(json: any): DataTypes.BankInfoType?
	if json == nil then return nil end
	return DataTypes.decodeJsonBankInfoType(json)
end
local function encodeOptionalBankInfoType(v: DataTypes.BankInfoType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonBankInfoType(v)
end
local function decodeOptionalBaseContactType(json: any): DataTypes.BaseContactType?
	if json == nil then return nil end
	return DataTypes.decodeJsonBaseContactType(json)
end
local function encodeOptionalBaseContactType(v: DataTypes.BaseContactType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonBaseContactType(v)
end
local function decodeOptionalBooleanType(json: any): DataTypes.BooleanType?
	if json == nil then return nil end
	return DataTypes.decodeJsonBooleanType(json)
end
local function encodeOptionalBooleanType(v: DataTypes.BooleanType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonBooleanType(v)
end
local function decodeOptionalBrokerageAccountType(json: any): DataTypes.BrokerageAccountType?
	if json == nil then return nil end
	return DataTypes.decodeJsonBrokerageAccountType(json)
end
local function encodeOptionalBrokerageAccountType(v: DataTypes.BrokerageAccountType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonBrokerageAccountType(v)
end
local function decodeOptionalBrokerageFirmType(json: any): DataTypes.BrokerageFirmType?
	if json == nil then return nil end
	return DataTypes.decodeJsonBrokerageFirmType(json)
end
local function encodeOptionalBrokerageFirmType(v: DataTypes.BrokerageFirmType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonBrokerageFirmType(v)
end
local function decodeOptionalCountryType(json: any): DataTypes.CountryType?
	if json == nil then return nil end
	return DataTypes.decodeJsonCountryType(json)
end
local function encodeOptionalCountryType(v: DataTypes.CountryType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonCountryType(v)
end
local function decodeOptionalCustomCompoundType(json: any): DataTypes.CustomCompoundType?
	if json == nil then return nil end
	return DataTypes.decodeJsonCustomCompoundType(json)
end
local function encodeOptionalCustomCompoundType(v: DataTypes.CustomCompoundType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonCustomCompoundType(v)
end
local function decodeOptionalDateTimeType(json: any): DataTypes.DateTimeType?
	if json == nil then return nil end
	return DataTypes.decodeJsonDateTimeType(json)
end
local function encodeOptionalDateTimeType(v: DataTypes.DateTimeType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonDateTimeType(v)
end
local function decodeOptionalEnumType(json: any): DataTypes.EnumType?
	if json == nil then return nil end
	return DataTypes.decodeJsonEnumType(json)
end
local function encodeOptionalEnumType(v: DataTypes.EnumType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonEnumType(v)
end
local function decodeOptionalIndividualNameType(json: any): DataTypes.IndividualNameType?
	if json == nil then return nil end
	return DataTypes.decodeJsonIndividualNameType(json)
end
local function encodeOptionalIndividualNameType(v: DataTypes.IndividualNameType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonIndividualNameType(v)
end
local function decodeOptionalMoneyType(json: any): DataTypes.MoneyType?
	if json == nil then return nil end
	return DataTypes.decodeJsonMoneyType(json)
end
local function encodeOptionalMoneyType(v: DataTypes.MoneyType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonMoneyType(v)
end
local function decodeOptionalMultipleCheckboxType(json: any): DataTypes.MultipleCheckboxType?
	if json == nil then return nil end
	return DataTypes.decodeJsonMultipleCheckboxType(json)
end
local function encodeOptionalMultipleCheckboxType(v: DataTypes.MultipleCheckboxType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonMultipleCheckboxType(v)
end
local function decodeOptionalNumberType(json: any): DataTypes.NumberType?
	if json == nil then return nil end
	return DataTypes.decodeJsonNumberType(json)
end
local function encodeOptionalNumberType(v: DataTypes.NumberType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonNumberType(v)
end
local function decodeOptionalPhoneFaxType(json: any): DataTypes.PhoneFaxType?
	if json == nil then return nil end
	return DataTypes.decodeJsonPhoneFaxType(json)
end
local function encodeOptionalPhoneFaxType(v: DataTypes.PhoneFaxType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonPhoneFaxType(v)
end
local function decodeOptionalRadioGroupType(json: any): DataTypes.RadioGroupType?
	if json == nil then return nil end
	return DataTypes.decodeJsonRadioGroupType(json)
end
local function encodeOptionalRadioGroupType(v: DataTypes.RadioGroupType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonRadioGroupType(v)
end
local function decodeOptionalServiceContactPointType(json: any): DataTypes.ServiceContactPointType?
	if json == nil then return nil end
	return DataTypes.decodeJsonServiceContactPointType(json)
end
local function encodeOptionalServiceContactPointType(v: DataTypes.ServiceContactPointType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonServiceContactPointType(v)
end
local function decodeOptionalSignatoryType(json: any): DataTypes.SignatoryType?
	if json == nil then return nil end
	return DataTypes.decodeJsonSignatoryType(json)
end
local function encodeOptionalSignatoryType(v: DataTypes.SignatoryType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonSignatoryType(v)
end
local function decodeOptionalStringType(json: any): DataTypes.StringType?
	if json == nil then return nil end
	return DataTypes.decodeJsonStringType(json)
end
local function encodeOptionalStringType(v: DataTypes.StringType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonStringType(v)
end
local function decodeOptionalSubmissionContactType(json: any): DataTypes.SubmissionContactType?
	if json == nil then return nil end
	return DataTypes.decodeJsonSubmissionContactType(json)
end
local function encodeOptionalSubmissionContactType(v: DataTypes.SubmissionContactType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonSubmissionContactType(v)
end
local function decodeOptionalWireInstructionsType(json: any): DataTypes.WireInstructionsType?
	if json == nil then return nil end
	return DataTypes.decodeJsonWireInstructionsType(json)
end
local function encodeOptionalWireInstructionsType(v: DataTypes.WireInstructionsType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonWireInstructionsType(v)
end

function TableSchema.encodeJsonFieldGroup(f: FieldGroup): {[string]: any}
	return {
		label = f.label,
		startIdx = f.startIdx,
		endIdx = f.endIdx,
	}
end

function TableSchema.decodeJsonFieldGroup(json: {[string]: any}): FieldGroup
	return {
		label = json.label or "",
		startIdx = json.startIdx or 0,
		endIdx = json.endIdx or 0,
	}
end

function TableSchema.encodeJsonSingleFieldType(si: SingleFieldType): {[string]: any}
	local result: {[string]: any} = {
		label = si.label,
	}
	result.stringType = encodeOptionalStringType(si.stringType)
	result.numberType = encodeOptionalNumberType(si.numberType)
	result.booleanType = encodeOptionalBooleanType(si.booleanType)
	result.enumType = encodeOptionalEnumType(si.enumType)
	result.multipleCheckboxType = encodeOptionalMultipleCheckboxType(si.multipleCheckboxType)
	result.radioGroupType = encodeOptionalRadioGroupType(si.radioGroupType)
	result.phoneFaxType = encodeOptionalPhoneFaxType(si.phoneFaxType)
	result.dateTimeType = encodeOptionalDateTimeType(si.dateTimeType)
	result.moneyType = encodeOptionalMoneyType(si.moneyType)
	result.addressType = encodeOptionalAddressType(si.addressType)
	result.individualNameType = encodeOptionalIndividualNameType(si.individualNameType)
	result.countryType = encodeOptionalCountryType(si.countryType)
	result.baseContactType = encodeOptionalBaseContactType(si.baseContactType)
	result.submissionContactType = encodeOptionalSubmissionContactType(si.submissionContactType)
	result.signatoryType = encodeOptionalSignatoryType(si.signatoryType)
	result.bankInfoType = encodeOptionalBankInfoType(si.bankInfoType)
	result.bankAccountInfoType = encodeOptionalBankAccountInfoType(si.bankAccountInfoType)
	result.wireInstructionsType = encodeOptionalWireInstructionsType(si.wireInstructionsType)
	result.brokerageFirmType = encodeOptionalBrokerageFirmType(si.brokerageFirmType)
	result.brokerageAccountType = encodeOptionalBrokerageAccountType(si.brokerageAccountType)
	result.serviceContactPointType = encodeOptionalServiceContactPointType(si.serviceContactPointType)
	result.customCompoundType = encodeOptionalCustomCompoundType(si.customCompoundType)
	return result
end

function TableSchema.decodeJsonSingleFieldType(json: {[string]: any}): SingleFieldType
	return {
		stringType = decodeOptionalStringType(json.stringType),
		numberType = decodeOptionalNumberType(json.numberType),
		booleanType = decodeOptionalBooleanType(json.booleanType),
		enumType = decodeOptionalEnumType(json.enumType),
		multipleCheckboxType = decodeOptionalMultipleCheckboxType(json.multipleCheckboxType),
		radioGroupType = decodeOptionalRadioGroupType(json.radioGroupType),
		phoneFaxType = decodeOptionalPhoneFaxType(json.phoneFaxType),
		dateTimeType = decodeOptionalDateTimeType(json.dateTimeType),
		moneyType = decodeOptionalMoneyType(json.moneyType),
		addressType = decodeOptionalAddressType(json.addressType),
		individualNameType = decodeOptionalIndividualNameType(json.individualNameType),
		countryType = decodeOptionalCountryType(json.countryType),
		baseContactType = decodeOptionalBaseContactType(json.baseContactType),
		submissionContactType = decodeOptionalSubmissionContactType(json.submissionContactType),
		signatoryType = decodeOptionalSignatoryType(json.signatoryType),
		bankInfoType = decodeOptionalBankInfoType(json.bankInfoType),
		bankAccountInfoType = decodeOptionalBankAccountInfoType(json.bankAccountInfoType),
		wireInstructionsType = decodeOptionalWireInstructionsType(json.wireInstructionsType),
		brokerageFirmType = decodeOptionalBrokerageFirmType(json.brokerageFirmType),
		brokerageAccountType = decodeOptionalBrokerageAccountType(json.brokerageAccountType),
		serviceContactPointType = decodeOptionalServiceContactPointType(json.serviceContactPointType),
		customCompoundType = decodeOptionalCustomCompoundType(json.customCompoundType),
		label = json.label or "",
	}
end

function TableSchema.encodeJsonTableSchema(ta: TableSchema): {[string]: any}
	local fieldsMapEncoded: {[string]: any} = {}
	for key, val in ta.fieldsMap do
		fieldsMapEncoded[key] = TableSchema.encodeJsonSingleFieldType(val)
	end
	local groupsEncoded: {any} = {}
	for _, item in ta.groups do
		table.insert(groupsEncoded, TableSchema.encodeJsonFieldGroup(item))
	end
	return {
		fieldKeysInOrder = ta.fieldKeysInOrder,
		label = ta.label,
		fieldsMap = fieldsMapEncoded,
		groups = groupsEncoded,
	}
end

function TableSchema.decodeJsonTableSchema(json: {[string]: any}): TableSchema
	local fieldsMap: {[string]: SingleFieldType} = {}
	if json.fieldsMap ~= nil then
		for key, val in json.fieldsMap :: {[string]: any} do
			fieldsMap[key] = TableSchema.decodeJsonSingleFieldType(val)
		end
	end
	local groups: {FieldGroup} = {}
	if json.groups ~= nil then
		for _, item in json.groups :: {{[string]: any}} do
			table.insert(groups, TableSchema.decodeJsonFieldGroup(item))
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
