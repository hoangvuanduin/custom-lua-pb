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

export type FieldValue = {
	typeId: string,
	[string]: any,
}

local function encodeOptionalStringField(field: StringType?): {[string]: any}?
	if field == nil then
		return nil
	end
	return DataTypes.encodeJsonStringType(field)
end

local function decodeOptionalStringField(json: any): StringType?
	if json == nil then
		return nil
	end
	return DataTypes.decodeJsonStringType(json)
end

local function encodeOptionalNumberField(field: NumberType?): {[string]: any}?
	if field == nil then
		return nil
	end
	return DataTypes.encodeJsonNumberType(field)
end

local function decodeOptionalNumberField(json: any): NumberType?
	if json == nil then
		return nil
	end
	return DataTypes.decodeJsonNumberType(json)
end

local function encodeOptionalMultipleCheckboxField(field: MultipleCheckboxType?): {[string]: any}?
	if field == nil then
		return nil
	end
	return DataTypes.encodeJsonMultipleCheckboxType(field)
end

local function decodeOptionalMultipleCheckboxField(json: any): MultipleCheckboxType?
	if json == nil then
		return nil
	end
	return DataTypes.decodeJsonMultipleCheckboxType(json)
end

export type AddressSubFields = {
	numberAndStreet: StringType?,
	city: StringType?,
	stateProvince: StringType?,
	country: StringType?,
	postalZipCode: StringType?,
	fullAddress: StringType?,
}
export type AddressType = {
	typeId: string,
	valueSubFields: AddressSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeAddressSubFields(config: {
	numberAndStreet: StringType?,
	city: StringType?,
	stateProvince: StringType?,
	country: StringType?,
	postalZipCode: StringType?,
	fullAddress: StringType?,
}?): AddressSubFields
	local c = config or {}
	return {
		numberAndStreet = c.numberAndStreet,
		city = c.city,
		stateProvince = c.stateProvince,
		country = c.country,
		postalZipCode = c.postalZipCode,
		fullAddress = c.fullAddress,
	}
end

function DataTypes.makeAddressType(config: {
	typeId: string?,
	valueSubFields: AddressSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): AddressType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonAddressSubFields(sf: AddressSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.numberAndStreet = encodeOptionalStringField(sf.numberAndStreet)
	result.city = encodeOptionalStringField(sf.city)
	result.stateProvince = encodeOptionalStringField(sf.stateProvince)
	result.country = encodeOptionalStringField(sf.country)
	result.postalZipCode = encodeOptionalStringField(sf.postalZipCode)
	result.fullAddress = encodeOptionalStringField(sf.fullAddress)
	return result
end

function DataTypes.decodeJsonAddressSubFields(json: {[string]: any}): AddressSubFields
	return {
		numberAndStreet = decodeOptionalStringField(json.numberAndStreet),
		city = decodeOptionalStringField(json.city),
		stateProvince = decodeOptionalStringField(json.stateProvince),
		country = decodeOptionalStringField(json.country),
		postalZipCode = decodeOptionalStringField(json.postalZipCode),
		fullAddress = decodeOptionalStringField(json.fullAddress),
	}
end

function DataTypes.encodeJsonAddressType(t: AddressType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonAddressSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonAddressType(json: {[string]: any}): AddressType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonAddressSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type BankAccountInfoSubFields = {
	bankName: StringType?,
	bankLocation: StringType?,
	aba: StringType?,
	swift: StringType?,
	accountName: StringType?,
	accountNumber: StringType?,
	iban: StringType?,
	furtherCreditName: StringType?,
	furtherCreditAccount: StringType?,
	correspondentBankAccount: StringType?,
}
export type BankAccountInfoType = {
	typeId: string,
	valueSubFields: BankAccountInfoSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeBankAccountInfoSubFields(config: {
	bankName: StringType?,
	bankLocation: StringType?,
	aba: StringType?,
	swift: StringType?,
	accountName: StringType?,
	accountNumber: StringType?,
	iban: StringType?,
	furtherCreditName: StringType?,
	furtherCreditAccount: StringType?,
	correspondentBankAccount: StringType?,
}?): BankAccountInfoSubFields
	local c = config or {}
	return {
		bankName = c.bankName,
		bankLocation = c.bankLocation,
		aba = c.aba,
		swift = c.swift,
		accountName = c.accountName,
		accountNumber = c.accountNumber,
		iban = c.iban,
		furtherCreditName = c.furtherCreditName,
		furtherCreditAccount = c.furtherCreditAccount,
		correspondentBankAccount = c.correspondentBankAccount,
	}
end

function DataTypes.makeBankAccountInfoType(config: {
	typeId: string?,
	valueSubFields: BankAccountInfoSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): BankAccountInfoType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonBankAccountInfoSubFields(sf: BankAccountInfoSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.bankName = encodeOptionalStringField(sf.bankName)
	result.bankLocation = encodeOptionalStringField(sf.bankLocation)
	result.aba = encodeOptionalStringField(sf.aba)
	result.swift = encodeOptionalStringField(sf.swift)
	result.accountName = encodeOptionalStringField(sf.accountName)
	result.accountNumber = encodeOptionalStringField(sf.accountNumber)
	result.iban = encodeOptionalStringField(sf.iban)
	result.furtherCreditName = encodeOptionalStringField(sf.furtherCreditName)
	result.furtherCreditAccount = encodeOptionalStringField(sf.furtherCreditAccount)
	result.correspondentBankAccount = encodeOptionalStringField(sf.correspondentBankAccount)
	return result
end

function DataTypes.decodeJsonBankAccountInfoSubFields(json: {[string]: any}): BankAccountInfoSubFields
	return {
		bankName = decodeOptionalStringField(json.bankName),
		bankLocation = decodeOptionalStringField(json.bankLocation),
		aba = decodeOptionalStringField(json.aba),
		swift = decodeOptionalStringField(json.swift),
		accountName = decodeOptionalStringField(json.accountName),
		accountNumber = decodeOptionalStringField(json.accountNumber),
		iban = decodeOptionalStringField(json.iban),
		furtherCreditName = decodeOptionalStringField(json.furtherCreditName),
		furtherCreditAccount = decodeOptionalStringField(json.furtherCreditAccount),
		correspondentBankAccount = decodeOptionalStringField(json.correspondentBankAccount),
	}
end

function DataTypes.encodeJsonBankAccountInfoType(t: BankAccountInfoType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonBankAccountInfoSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonBankAccountInfoType(json: {[string]: any}): BankAccountInfoType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonBankAccountInfoSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type BankInfoSubFields = {
	bankName: StringType?,
	bankLocation: StringType?,
	aba: StringType?,
	swift: StringType?,
}
export type BankInfoType = {
	typeId: string,
	valueSubFields: BankInfoSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeBankInfoSubFields(config: {
	bankName: StringType?,
	bankLocation: StringType?,
	aba: StringType?,
	swift: StringType?,
}?): BankInfoSubFields
	local c = config or {}
	return {
		bankName = c.bankName,
		bankLocation = c.bankLocation,
		aba = c.aba,
		swift = c.swift,
	}
end

function DataTypes.makeBankInfoType(config: {
	typeId: string?,
	valueSubFields: BankInfoSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): BankInfoType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonBankInfoSubFields(sf: BankInfoSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.bankName = encodeOptionalStringField(sf.bankName)
	result.bankLocation = encodeOptionalStringField(sf.bankLocation)
	result.aba = encodeOptionalStringField(sf.aba)
	result.swift = encodeOptionalStringField(sf.swift)
	return result
end

function DataTypes.decodeJsonBankInfoSubFields(json: {[string]: any}): BankInfoSubFields
	return {
		bankName = decodeOptionalStringField(json.bankName),
		bankLocation = decodeOptionalStringField(json.bankLocation),
		aba = decodeOptionalStringField(json.aba),
		swift = decodeOptionalStringField(json.swift),
	}
end

function DataTypes.encodeJsonBankInfoType(t: BankInfoType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonBankInfoSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonBankInfoType(json: {[string]: any}): BankInfoType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonBankInfoSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type BaseContactSubFields = {
	prefixName: StringType?,
	firstName: StringType?,
	middleName: StringType?,
	lastName: StringType?,
	suffixName: StringType?,
	fullName: StringType?,
	preferredName: StringType?,
	jobTitle: StringType?,
	email: StringType?,
	phone: StringType?,
	fax: StringType?,
}
export type BaseContactType = {
	typeId: string,
	valueSubFields: BaseContactSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeBaseContactSubFields(config: {
	prefixName: StringType?,
	firstName: StringType?,
	middleName: StringType?,
	lastName: StringType?,
	suffixName: StringType?,
	fullName: StringType?,
	preferredName: StringType?,
	jobTitle: StringType?,
	email: StringType?,
	phone: StringType?,
	fax: StringType?,
}?): BaseContactSubFields
	local c = config or {}
	return {
		prefixName = c.prefixName,
		firstName = c.firstName,
		middleName = c.middleName,
		lastName = c.lastName,
		suffixName = c.suffixName,
		fullName = c.fullName,
		preferredName = c.preferredName,
		jobTitle = c.jobTitle,
		email = c.email,
		phone = c.phone,
		fax = c.fax,
	}
end

function DataTypes.makeBaseContactType(config: {
	typeId: string?,
	valueSubFields: BaseContactSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): BaseContactType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonBaseContactSubFields(sf: BaseContactSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.prefixName = encodeOptionalStringField(sf.prefixName)
	result.firstName = encodeOptionalStringField(sf.firstName)
	result.middleName = encodeOptionalStringField(sf.middleName)
	result.lastName = encodeOptionalStringField(sf.lastName)
	result.suffixName = encodeOptionalStringField(sf.suffixName)
	result.fullName = encodeOptionalStringField(sf.fullName)
	result.preferredName = encodeOptionalStringField(sf.preferredName)
	result.jobTitle = encodeOptionalStringField(sf.jobTitle)
	result.email = encodeOptionalStringField(sf.email)
	result.phone = encodeOptionalStringField(sf.phone)
	result.fax = encodeOptionalStringField(sf.fax)
	return result
end

function DataTypes.decodeJsonBaseContactSubFields(json: {[string]: any}): BaseContactSubFields
	return {
		prefixName = decodeOptionalStringField(json.prefixName),
		firstName = decodeOptionalStringField(json.firstName),
		middleName = decodeOptionalStringField(json.middleName),
		lastName = decodeOptionalStringField(json.lastName),
		suffixName = decodeOptionalStringField(json.suffixName),
		fullName = decodeOptionalStringField(json.fullName),
		preferredName = decodeOptionalStringField(json.preferredName),
		jobTitle = decodeOptionalStringField(json.jobTitle),
		email = decodeOptionalStringField(json.email),
		phone = decodeOptionalStringField(json.phone),
		fax = decodeOptionalStringField(json.fax),
	}
end

function DataTypes.encodeJsonBaseContactType(t: BaseContactType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonBaseContactSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonBaseContactType(json: {[string]: any}): BaseContactType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonBaseContactSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type BrokerageAccountSubFields = {
	firmName: StringType?,
	firmOfficeBranchLocation: StringType?,
	dtcNumber: StringType?,
	crestParticipantId: StringType?,
	bicCode: StringType?,
	beneficiaryAccountNumber: StringType?,
	beneficiaryAccountName: StringType?,
	intermediaryAccountName: StringType?,
	intermediaryAccountNumber: StringType?,
	intermediaryBicCode: StringType?,
	securitiesSettlementInstructions: StringType?,
	attention: StringType?,
}
export type BrokerageAccountType = {
	typeId: string,
	valueSubFields: BrokerageAccountSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeBrokerageAccountSubFields(config: {
	firmName: StringType?,
	firmOfficeBranchLocation: StringType?,
	dtcNumber: StringType?,
	crestParticipantId: StringType?,
	bicCode: StringType?,
	beneficiaryAccountNumber: StringType?,
	beneficiaryAccountName: StringType?,
	intermediaryAccountName: StringType?,
	intermediaryAccountNumber: StringType?,
	intermediaryBicCode: StringType?,
	securitiesSettlementInstructions: StringType?,
	attention: StringType?,
}?): BrokerageAccountSubFields
	local c = config or {}
	return {
		firmName = c.firmName,
		firmOfficeBranchLocation = c.firmOfficeBranchLocation,
		dtcNumber = c.dtcNumber,
		crestParticipantId = c.crestParticipantId,
		bicCode = c.bicCode,
		beneficiaryAccountNumber = c.beneficiaryAccountNumber,
		beneficiaryAccountName = c.beneficiaryAccountName,
		intermediaryAccountName = c.intermediaryAccountName,
		intermediaryAccountNumber = c.intermediaryAccountNumber,
		intermediaryBicCode = c.intermediaryBicCode,
		securitiesSettlementInstructions = c.securitiesSettlementInstructions,
		attention = c.attention,
	}
end

function DataTypes.makeBrokerageAccountType(config: {
	typeId: string?,
	valueSubFields: BrokerageAccountSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): BrokerageAccountType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonBrokerageAccountSubFields(sf: BrokerageAccountSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.firmName = encodeOptionalStringField(sf.firmName)
	result.firmOfficeBranchLocation = encodeOptionalStringField(sf.firmOfficeBranchLocation)
	result.dtcNumber = encodeOptionalStringField(sf.dtcNumber)
	result.crestParticipantId = encodeOptionalStringField(sf.crestParticipantId)
	result.bicCode = encodeOptionalStringField(sf.bicCode)
	result.beneficiaryAccountNumber = encodeOptionalStringField(sf.beneficiaryAccountNumber)
	result.beneficiaryAccountName = encodeOptionalStringField(sf.beneficiaryAccountName)
	result.intermediaryAccountName = encodeOptionalStringField(sf.intermediaryAccountName)
	result.intermediaryAccountNumber = encodeOptionalStringField(sf.intermediaryAccountNumber)
	result.intermediaryBicCode = encodeOptionalStringField(sf.intermediaryBicCode)
	result.securitiesSettlementInstructions = encodeOptionalStringField(sf.securitiesSettlementInstructions)
	result.attention = encodeOptionalStringField(sf.attention)
	return result
end

function DataTypes.decodeJsonBrokerageAccountSubFields(json: {[string]: any}): BrokerageAccountSubFields
	return {
		firmName = decodeOptionalStringField(json.firmName),
		firmOfficeBranchLocation = decodeOptionalStringField(json.firmOfficeBranchLocation),
		dtcNumber = decodeOptionalStringField(json.dtcNumber),
		crestParticipantId = decodeOptionalStringField(json.crestParticipantId),
		bicCode = decodeOptionalStringField(json.bicCode),
		beneficiaryAccountNumber = decodeOptionalStringField(json.beneficiaryAccountNumber),
		beneficiaryAccountName = decodeOptionalStringField(json.beneficiaryAccountName),
		intermediaryAccountName = decodeOptionalStringField(json.intermediaryAccountName),
		intermediaryAccountNumber = decodeOptionalStringField(json.intermediaryAccountNumber),
		intermediaryBicCode = decodeOptionalStringField(json.intermediaryBicCode),
		securitiesSettlementInstructions = decodeOptionalStringField(json.securitiesSettlementInstructions),
		attention = decodeOptionalStringField(json.attention),
	}
end

function DataTypes.encodeJsonBrokerageAccountType(t: BrokerageAccountType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonBrokerageAccountSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonBrokerageAccountType(json: {[string]: any}): BrokerageAccountType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonBrokerageAccountSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type BrokerageFirmSubFields = {
	firmName: StringType?,
	firmOfficeBranchLocation: StringType?,
}
export type BrokerageFirmType = {
	typeId: string,
	valueSubFields: BrokerageFirmSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeBrokerageFirmSubFields(config: {
	firmName: StringType?,
	firmOfficeBranchLocation: StringType?,
}?): BrokerageFirmSubFields
	local c = config or {}
	return {
		firmName = c.firmName,
		firmOfficeBranchLocation = c.firmOfficeBranchLocation,
	}
end

function DataTypes.makeBrokerageFirmType(config: {
	typeId: string?,
	valueSubFields: BrokerageFirmSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): BrokerageFirmType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonBrokerageFirmSubFields(sf: BrokerageFirmSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.firmName = encodeOptionalStringField(sf.firmName)
	result.firmOfficeBranchLocation = encodeOptionalStringField(sf.firmOfficeBranchLocation)
	return result
end

function DataTypes.decodeJsonBrokerageFirmSubFields(json: {[string]: any}): BrokerageFirmSubFields
	return {
		firmName = decodeOptionalStringField(json.firmName),
		firmOfficeBranchLocation = decodeOptionalStringField(json.firmOfficeBranchLocation),
	}
end

function DataTypes.encodeJsonBrokerageFirmType(t: BrokerageFirmType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonBrokerageFirmSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonBrokerageFirmType(json: {[string]: any}): BrokerageFirmType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonBrokerageFirmSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type CountrySubFields = {
	countryName: StringType?,
	isoCountryCode: StringType?,
}
export type CountryType = {
	typeId: string,
	valueSubFields: CountrySubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeCountrySubFields(config: {
	countryName: StringType?,
	isoCountryCode: StringType?,
}?): CountrySubFields
	local c = config or {}
	return {
		countryName = c.countryName,
		isoCountryCode = c.isoCountryCode,
	}
end

function DataTypes.makeCountryType(config: {
	typeId: string?,
	valueSubFields: CountrySubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): CountryType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonCountrySubFields(sf: CountrySubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.countryName = encodeOptionalStringField(sf.countryName)
	result.isoCountryCode = encodeOptionalStringField(sf.isoCountryCode)
	return result
end

function DataTypes.decodeJsonCountrySubFields(json: {[string]: any}): CountrySubFields
	return {
		countryName = decodeOptionalStringField(json.countryName),
		isoCountryCode = decodeOptionalStringField(json.isoCountryCode),
	}
end

function DataTypes.encodeJsonCountryType(t: CountryType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonCountrySubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonCountryType(json: {[string]: any}): CountryType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonCountrySubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type DateTimeSubFields = {
	epochMs: NumberType?,
	zoneOffset: StringType?,
}
export type DateTimeType = {
	typeId: string,
	valueSubFields: DateTimeSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeDateTimeSubFields(config: {
	epochMs: NumberType?,
	zoneOffset: StringType?,
}?): DateTimeSubFields
	local c = config or {}
	return {
		epochMs = c.epochMs,
		zoneOffset = c.zoneOffset,
	}
end

function DataTypes.makeDateTimeType(config: {
	typeId: string?,
	valueSubFields: DateTimeSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): DateTimeType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonDateTimeSubFields(sf: DateTimeSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.epochMs = encodeOptionalNumberField(sf.epochMs)
	result.zoneOffset = encodeOptionalStringField(sf.zoneOffset)
	return result
end

function DataTypes.decodeJsonDateTimeSubFields(json: {[string]: any}): DateTimeSubFields
	return {
		epochMs = decodeOptionalNumberField(json.epochMs),
		zoneOffset = decodeOptionalStringField(json.zoneOffset),
	}
end

function DataTypes.encodeJsonDateTimeType(t: DateTimeType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonDateTimeSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonDateTimeType(json: {[string]: any}): DateTimeType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonDateTimeSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type IndividualNameSubFields = {
	prefixName: StringType?,
	firstName: StringType?,
	middleName: StringType?,
	lastName: StringType?,
	suffixName: StringType?,
	fullName: StringType?,
	preferredName: StringType?,
	jobTitle: StringType?,
}
export type IndividualNameType = {
	typeId: string,
	valueSubFields: IndividualNameSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeIndividualNameSubFields(config: {
	prefixName: StringType?,
	firstName: StringType?,
	middleName: StringType?,
	lastName: StringType?,
	suffixName: StringType?,
	fullName: StringType?,
	preferredName: StringType?,
	jobTitle: StringType?,
}?): IndividualNameSubFields
	local c = config or {}
	return {
		prefixName = c.prefixName,
		firstName = c.firstName,
		middleName = c.middleName,
		lastName = c.lastName,
		suffixName = c.suffixName,
		fullName = c.fullName,
		preferredName = c.preferredName,
		jobTitle = c.jobTitle,
	}
end

function DataTypes.makeIndividualNameType(config: {
	typeId: string?,
	valueSubFields: IndividualNameSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): IndividualNameType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonIndividualNameSubFields(sf: IndividualNameSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.prefixName = encodeOptionalStringField(sf.prefixName)
	result.firstName = encodeOptionalStringField(sf.firstName)
	result.middleName = encodeOptionalStringField(sf.middleName)
	result.lastName = encodeOptionalStringField(sf.lastName)
	result.suffixName = encodeOptionalStringField(sf.suffixName)
	result.fullName = encodeOptionalStringField(sf.fullName)
	result.preferredName = encodeOptionalStringField(sf.preferredName)
	result.jobTitle = encodeOptionalStringField(sf.jobTitle)
	return result
end

function DataTypes.decodeJsonIndividualNameSubFields(json: {[string]: any}): IndividualNameSubFields
	return {
		prefixName = decodeOptionalStringField(json.prefixName),
		firstName = decodeOptionalStringField(json.firstName),
		middleName = decodeOptionalStringField(json.middleName),
		lastName = decodeOptionalStringField(json.lastName),
		suffixName = decodeOptionalStringField(json.suffixName),
		fullName = decodeOptionalStringField(json.fullName),
		preferredName = decodeOptionalStringField(json.preferredName),
		jobTitle = decodeOptionalStringField(json.jobTitle),
	}
end

function DataTypes.encodeJsonIndividualNameType(t: IndividualNameType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonIndividualNameSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonIndividualNameType(json: {[string]: any}): IndividualNameType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonIndividualNameSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type MoneySubFields = {
	amount: NumberType?,
	isoCurrencyCode: StringType?,
}
export type MoneyType = {
	typeId: string,
	valueSubFields: MoneySubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeMoneySubFields(config: {
	amount: NumberType?,
	isoCurrencyCode: StringType?,
}?): MoneySubFields
	local c = config or {}
	return {
		amount = c.amount,
		isoCurrencyCode = c.isoCurrencyCode,
	}
end

function DataTypes.makeMoneyType(config: {
	typeId: string?,
	valueSubFields: MoneySubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): MoneyType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonMoneySubFields(sf: MoneySubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.amount = encodeOptionalNumberField(sf.amount)
	result.isoCurrencyCode = encodeOptionalStringField(sf.isoCurrencyCode)
	return result
end

function DataTypes.decodeJsonMoneySubFields(json: {[string]: any}): MoneySubFields
	return {
		amount = decodeOptionalNumberField(json.amount),
		isoCurrencyCode = decodeOptionalStringField(json.isoCurrencyCode),
	}
end

function DataTypes.encodeJsonMoneyType(t: MoneyType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonMoneySubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonMoneyType(json: {[string]: any}): MoneyType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonMoneySubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type PhoneFaxSubFields = {
	countryCode: StringType?,
	areaCode: StringType?,
	subscriberNumber: StringType?,
}
export type PhoneFaxType = {
	typeId: string,
	valueSubFields: PhoneFaxSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makePhoneFaxSubFields(config: {
	countryCode: StringType?,
	areaCode: StringType?,
	subscriberNumber: StringType?,
}?): PhoneFaxSubFields
	local c = config or {}
	return {
		countryCode = c.countryCode,
		areaCode = c.areaCode,
		subscriberNumber = c.subscriberNumber,
	}
end

function DataTypes.makePhoneFaxType(config: {
	typeId: string?,
	valueSubFields: PhoneFaxSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): PhoneFaxType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonPhoneFaxSubFields(sf: PhoneFaxSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.countryCode = encodeOptionalStringField(sf.countryCode)
	result.areaCode = encodeOptionalStringField(sf.areaCode)
	result.subscriberNumber = encodeOptionalStringField(sf.subscriberNumber)
	return result
end

function DataTypes.decodeJsonPhoneFaxSubFields(json: {[string]: any}): PhoneFaxSubFields
	return {
		countryCode = decodeOptionalStringField(json.countryCode),
		areaCode = decodeOptionalStringField(json.areaCode),
		subscriberNumber = decodeOptionalStringField(json.subscriberNumber),
	}
end

function DataTypes.encodeJsonPhoneFaxType(t: PhoneFaxType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonPhoneFaxSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonPhoneFaxType(json: {[string]: any}): PhoneFaxType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonPhoneFaxSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type ServiceContactPointSubFields = {
	name: StringType?,
	officeBranchLocation: StringType?,
	phone: StringType?,
	email: StringType?,
}
export type ServiceContactPointType = {
	typeId: string,
	valueSubFields: ServiceContactPointSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeServiceContactPointSubFields(config: {
	name: StringType?,
	officeBranchLocation: StringType?,
	phone: StringType?,
	email: StringType?,
}?): ServiceContactPointSubFields
	local c = config or {}
	return {
		name = c.name,
		officeBranchLocation = c.officeBranchLocation,
		phone = c.phone,
		email = c.email,
	}
end

function DataTypes.makeServiceContactPointType(config: {
	typeId: string?,
	valueSubFields: ServiceContactPointSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): ServiceContactPointType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonServiceContactPointSubFields(sf: ServiceContactPointSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.name = encodeOptionalStringField(sf.name)
	result.officeBranchLocation = encodeOptionalStringField(sf.officeBranchLocation)
	result.phone = encodeOptionalStringField(sf.phone)
	result.email = encodeOptionalStringField(sf.email)
	return result
end

function DataTypes.decodeJsonServiceContactPointSubFields(json: {[string]: any}): ServiceContactPointSubFields
	return {
		name = decodeOptionalStringField(json.name),
		officeBranchLocation = decodeOptionalStringField(json.officeBranchLocation),
		phone = decodeOptionalStringField(json.phone),
		email = decodeOptionalStringField(json.email),
	}
end

function DataTypes.encodeJsonServiceContactPointType(t: ServiceContactPointType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonServiceContactPointSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonServiceContactPointType(json: {[string]: any}): ServiceContactPointType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonServiceContactPointSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type SignatorySubFields = {
	numberAndStreet: StringType?,
	city: StringType?,
	stateProvince: StringType?,
	country: StringType?,
	postalZipCode: StringType?,
	fullAddress: StringType?,
	name: StringType?,
	date: StringType?,
	title: StringType?,
	phone: StringType?,
	email: StringType?,
	signerRole: StringType?,
}
export type SignatoryType = {
	typeId: string,
	valueSubFields: SignatorySubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeSignatorySubFields(config: {
	numberAndStreet: StringType?,
	city: StringType?,
	stateProvince: StringType?,
	country: StringType?,
	postalZipCode: StringType?,
	fullAddress: StringType?,
	name: StringType?,
	date: StringType?,
	title: StringType?,
	phone: StringType?,
	email: StringType?,
	signerRole: StringType?,
}?): SignatorySubFields
	local c = config or {}
	return {
		numberAndStreet = c.numberAndStreet,
		city = c.city,
		stateProvince = c.stateProvince,
		country = c.country,
		postalZipCode = c.postalZipCode,
		fullAddress = c.fullAddress,
		name = c.name,
		date = c.date,
		title = c.title,
		phone = c.phone,
		email = c.email,
		signerRole = c.signerRole,
	}
end

function DataTypes.makeSignatoryType(config: {
	typeId: string?,
	valueSubFields: SignatorySubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): SignatoryType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonSignatorySubFields(sf: SignatorySubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.numberAndStreet = encodeOptionalStringField(sf.numberAndStreet)
	result.city = encodeOptionalStringField(sf.city)
	result.stateProvince = encodeOptionalStringField(sf.stateProvince)
	result.country = encodeOptionalStringField(sf.country)
	result.postalZipCode = encodeOptionalStringField(sf.postalZipCode)
	result.fullAddress = encodeOptionalStringField(sf.fullAddress)
	result.name = encodeOptionalStringField(sf.name)
	result.date = encodeOptionalStringField(sf.date)
	result.title = encodeOptionalStringField(sf.title)
	result.phone = encodeOptionalStringField(sf.phone)
	result.email = encodeOptionalStringField(sf.email)
	result.signerRole = encodeOptionalStringField(sf.signerRole)
	return result
end

function DataTypes.decodeJsonSignatorySubFields(json: {[string]: any}): SignatorySubFields
	return {
		numberAndStreet = decodeOptionalStringField(json.numberAndStreet),
		city = decodeOptionalStringField(json.city),
		stateProvince = decodeOptionalStringField(json.stateProvince),
		country = decodeOptionalStringField(json.country),
		postalZipCode = decodeOptionalStringField(json.postalZipCode),
		fullAddress = decodeOptionalStringField(json.fullAddress),
		name = decodeOptionalStringField(json.name),
		date = decodeOptionalStringField(json.date),
		title = decodeOptionalStringField(json.title),
		phone = decodeOptionalStringField(json.phone),
		email = decodeOptionalStringField(json.email),
		signerRole = decodeOptionalStringField(json.signerRole),
	}
end

function DataTypes.encodeJsonSignatoryType(t: SignatoryType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonSignatorySubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonSignatoryType(json: {[string]: any}): SignatoryType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonSignatorySubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type SubmissionContactSubFields = {
	prefixName: StringType?,
	firstName: StringType?,
	middleName: StringType?,
	lastName: StringType?,
	suffixName: StringType?,
	fullName: StringType?,
	preferredName: StringType?,
	jobTitle: StringType?,
	email: StringType?,
	phone: StringType?,
	fax: StringType?,
	typeOfCorrespondences: MultipleCheckboxType?,
	contactTypes: MultipleCheckboxType?,
	contactRoles: MultipleCheckboxType?,
	customId: StringType?,
	addressNumberAndStreet: StringType?,
	addressCity: StringType?,
	addressStateProvince: StringType?,
	addressCountry: StringType?,
	addressPostalZipCode: StringType?,
	addressFullAddress: StringType?,
	relationshipWithPrimary: StringType?,
	occupation: StringType?,
	companyName: StringType?,
	companyAddressNumberAndStreet: StringType?,
	companyAddressCity: StringType?,
	companyAddressStateProvince: StringType?,
	companyAddressCountry: StringType?,
	companyAddressPostalZipCode: StringType?,
	companyAddressFullAddress: StringType?,
	specifiedCompanyOrgUnit: StringType?,
	individualEmail: StringType?,
	individualPhone: StringType?,
	individualFax: StringType?,
}
export type SubmissionContactType = {
	typeId: string,
	valueSubFields: SubmissionContactSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeSubmissionContactSubFields(config: {
	prefixName: StringType?,
	firstName: StringType?,
	middleName: StringType?,
	lastName: StringType?,
	suffixName: StringType?,
	fullName: StringType?,
	preferredName: StringType?,
	jobTitle: StringType?,
	email: StringType?,
	phone: StringType?,
	fax: StringType?,
	typeOfCorrespondences: MultipleCheckboxType?,
	contactTypes: MultipleCheckboxType?,
	contactRoles: MultipleCheckboxType?,
	customId: StringType?,
	addressNumberAndStreet: StringType?,
	addressCity: StringType?,
	addressStateProvince: StringType?,
	addressCountry: StringType?,
	addressPostalZipCode: StringType?,
	addressFullAddress: StringType?,
	relationshipWithPrimary: StringType?,
	occupation: StringType?,
	companyName: StringType?,
	companyAddressNumberAndStreet: StringType?,
	companyAddressCity: StringType?,
	companyAddressStateProvince: StringType?,
	companyAddressCountry: StringType?,
	companyAddressPostalZipCode: StringType?,
	companyAddressFullAddress: StringType?,
	specifiedCompanyOrgUnit: StringType?,
	individualEmail: StringType?,
	individualPhone: StringType?,
	individualFax: StringType?,
}?): SubmissionContactSubFields
	local c = config or {}
	return {
		prefixName = c.prefixName,
		firstName = c.firstName,
		middleName = c.middleName,
		lastName = c.lastName,
		suffixName = c.suffixName,
		fullName = c.fullName,
		preferredName = c.preferredName,
		jobTitle = c.jobTitle,
		email = c.email,
		phone = c.phone,
		fax = c.fax,
		typeOfCorrespondences = c.typeOfCorrespondences,
		contactTypes = c.contactTypes,
		contactRoles = c.contactRoles,
		customId = c.customId,
		addressNumberAndStreet = c.addressNumberAndStreet,
		addressCity = c.addressCity,
		addressStateProvince = c.addressStateProvince,
		addressCountry = c.addressCountry,
		addressPostalZipCode = c.addressPostalZipCode,
		addressFullAddress = c.addressFullAddress,
		relationshipWithPrimary = c.relationshipWithPrimary,
		occupation = c.occupation,
		companyName = c.companyName,
		companyAddressNumberAndStreet = c.companyAddressNumberAndStreet,
		companyAddressCity = c.companyAddressCity,
		companyAddressStateProvince = c.companyAddressStateProvince,
		companyAddressCountry = c.companyAddressCountry,
		companyAddressPostalZipCode = c.companyAddressPostalZipCode,
		companyAddressFullAddress = c.companyAddressFullAddress,
		specifiedCompanyOrgUnit = c.specifiedCompanyOrgUnit,
		individualEmail = c.individualEmail,
		individualPhone = c.individualPhone,
		individualFax = c.individualFax,
	}
end

function DataTypes.makeSubmissionContactType(config: {
	typeId: string?,
	valueSubFields: SubmissionContactSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): SubmissionContactType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonSubmissionContactSubFields(sf: SubmissionContactSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.prefixName = encodeOptionalStringField(sf.prefixName)
	result.firstName = encodeOptionalStringField(sf.firstName)
	result.middleName = encodeOptionalStringField(sf.middleName)
	result.lastName = encodeOptionalStringField(sf.lastName)
	result.suffixName = encodeOptionalStringField(sf.suffixName)
	result.fullName = encodeOptionalStringField(sf.fullName)
	result.preferredName = encodeOptionalStringField(sf.preferredName)
	result.jobTitle = encodeOptionalStringField(sf.jobTitle)
	result.email = encodeOptionalStringField(sf.email)
	result.phone = encodeOptionalStringField(sf.phone)
	result.fax = encodeOptionalStringField(sf.fax)
	result.typeOfCorrespondences = encodeOptionalMultipleCheckboxField(sf.typeOfCorrespondences)
	result.contactTypes = encodeOptionalMultipleCheckboxField(sf.contactTypes)
	result.contactRoles = encodeOptionalMultipleCheckboxField(sf.contactRoles)
	result.customId = encodeOptionalStringField(sf.customId)
	result.addressNumberAndStreet = encodeOptionalStringField(sf.addressNumberAndStreet)
	result.addressCity = encodeOptionalStringField(sf.addressCity)
	result.addressStateProvince = encodeOptionalStringField(sf.addressStateProvince)
	result.addressCountry = encodeOptionalStringField(sf.addressCountry)
	result.addressPostalZipCode = encodeOptionalStringField(sf.addressPostalZipCode)
	result.addressFullAddress = encodeOptionalStringField(sf.addressFullAddress)
	result.relationshipWithPrimary = encodeOptionalStringField(sf.relationshipWithPrimary)
	result.occupation = encodeOptionalStringField(sf.occupation)
	result.companyName = encodeOptionalStringField(sf.companyName)
	result.companyAddressNumberAndStreet = encodeOptionalStringField(sf.companyAddressNumberAndStreet)
	result.companyAddressCity = encodeOptionalStringField(sf.companyAddressCity)
	result.companyAddressStateProvince = encodeOptionalStringField(sf.companyAddressStateProvince)
	result.companyAddressCountry = encodeOptionalStringField(sf.companyAddressCountry)
	result.companyAddressPostalZipCode = encodeOptionalStringField(sf.companyAddressPostalZipCode)
	result.companyAddressFullAddress = encodeOptionalStringField(sf.companyAddressFullAddress)
	result.specifiedCompanyOrgUnit = encodeOptionalStringField(sf.specifiedCompanyOrgUnit)
	result.individualEmail = encodeOptionalStringField(sf.individualEmail)
	result.individualPhone = encodeOptionalStringField(sf.individualPhone)
	result.individualFax = encodeOptionalStringField(sf.individualFax)
	return result
end

function DataTypes.decodeJsonSubmissionContactSubFields(json: {[string]: any}): SubmissionContactSubFields
	return {
		prefixName = decodeOptionalStringField(json.prefixName),
		firstName = decodeOptionalStringField(json.firstName),
		middleName = decodeOptionalStringField(json.middleName),
		lastName = decodeOptionalStringField(json.lastName),
		suffixName = decodeOptionalStringField(json.suffixName),
		fullName = decodeOptionalStringField(json.fullName),
		preferredName = decodeOptionalStringField(json.preferredName),
		jobTitle = decodeOptionalStringField(json.jobTitle),
		email = decodeOptionalStringField(json.email),
		phone = decodeOptionalStringField(json.phone),
		fax = decodeOptionalStringField(json.fax),
		typeOfCorrespondences = decodeOptionalMultipleCheckboxField(json.typeOfCorrespondences),
		contactTypes = decodeOptionalMultipleCheckboxField(json.contactTypes),
		contactRoles = decodeOptionalMultipleCheckboxField(json.contactRoles),
		customId = decodeOptionalStringField(json.customId),
		addressNumberAndStreet = decodeOptionalStringField(json.addressNumberAndStreet),
		addressCity = decodeOptionalStringField(json.addressCity),
		addressStateProvince = decodeOptionalStringField(json.addressStateProvince),
		addressCountry = decodeOptionalStringField(json.addressCountry),
		addressPostalZipCode = decodeOptionalStringField(json.addressPostalZipCode),
		addressFullAddress = decodeOptionalStringField(json.addressFullAddress),
		relationshipWithPrimary = decodeOptionalStringField(json.relationshipWithPrimary),
		occupation = decodeOptionalStringField(json.occupation),
		companyName = decodeOptionalStringField(json.companyName),
		companyAddressNumberAndStreet = decodeOptionalStringField(json.companyAddressNumberAndStreet),
		companyAddressCity = decodeOptionalStringField(json.companyAddressCity),
		companyAddressStateProvince = decodeOptionalStringField(json.companyAddressStateProvince),
		companyAddressCountry = decodeOptionalStringField(json.companyAddressCountry),
		companyAddressPostalZipCode = decodeOptionalStringField(json.companyAddressPostalZipCode),
		companyAddressFullAddress = decodeOptionalStringField(json.companyAddressFullAddress),
		specifiedCompanyOrgUnit = decodeOptionalStringField(json.specifiedCompanyOrgUnit),
		individualEmail = decodeOptionalStringField(json.individualEmail),
		individualPhone = decodeOptionalStringField(json.individualPhone),
		individualFax = decodeOptionalStringField(json.individualFax),
	}
end

function DataTypes.encodeJsonSubmissionContactType(t: SubmissionContactType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonSubmissionContactSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonSubmissionContactType(json: {[string]: any}): SubmissionContactType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonSubmissionContactSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type WireInstructionsSubFields = {
	bankName: StringType?,
	bankLocation: StringType?,
	aba: StringType?,
	swift: StringType?,
	accountName: StringType?,
	accountNumber: StringType?,
	iban: StringType?,
	furtherCreditName: StringType?,
	furtherCreditAccount: StringType?,
	correspondentBankAccount: StringType?,
	attention: StringType?,
}
export type WireInstructionsType = {
	typeId: string,
	valueSubFields: WireInstructionsSubFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeWireInstructionsSubFields(config: {
	bankName: StringType?,
	bankLocation: StringType?,
	aba: StringType?,
	swift: StringType?,
	accountName: StringType?,
	accountNumber: StringType?,
	iban: StringType?,
	furtherCreditName: StringType?,
	furtherCreditAccount: StringType?,
	correspondentBankAccount: StringType?,
	attention: StringType?,
}?): WireInstructionsSubFields
	local c = config or {}
	return {
		bankName = c.bankName,
		bankLocation = c.bankLocation,
		aba = c.aba,
		swift = c.swift,
		accountName = c.accountName,
		accountNumber = c.accountNumber,
		iban = c.iban,
		furtherCreditName = c.furtherCreditName,
		furtherCreditAccount = c.furtherCreditAccount,
		correspondentBankAccount = c.correspondentBankAccount,
		attention = c.attention,
	}
end

function DataTypes.makeWireInstructionsType(config: {
	typeId: string?,
	valueSubFields: WireInstructionsSubFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): WireInstructionsType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function DataTypes.encodeJsonWireInstructionsSubFields(sf: WireInstructionsSubFields): {[string]: any}
	local result: {[string]: any} = {}
	result.bankName = encodeOptionalStringField(sf.bankName)
	result.bankLocation = encodeOptionalStringField(sf.bankLocation)
	result.aba = encodeOptionalStringField(sf.aba)
	result.swift = encodeOptionalStringField(sf.swift)
	result.accountName = encodeOptionalStringField(sf.accountName)
	result.accountNumber = encodeOptionalStringField(sf.accountNumber)
	result.iban = encodeOptionalStringField(sf.iban)
	result.furtherCreditName = encodeOptionalStringField(sf.furtherCreditName)
	result.furtherCreditAccount = encodeOptionalStringField(sf.furtherCreditAccount)
	result.correspondentBankAccount = encodeOptionalStringField(sf.correspondentBankAccount)
	result.attention = encodeOptionalStringField(sf.attention)
	return result
end

function DataTypes.decodeJsonWireInstructionsSubFields(json: {[string]: any}): WireInstructionsSubFields
	return {
		bankName = decodeOptionalStringField(json.bankName),
		bankLocation = decodeOptionalStringField(json.bankLocation),
		aba = decodeOptionalStringField(json.aba),
		swift = decodeOptionalStringField(json.swift),
		accountName = decodeOptionalStringField(json.accountName),
		accountNumber = decodeOptionalStringField(json.accountNumber),
		iban = decodeOptionalStringField(json.iban),
		furtherCreditName = decodeOptionalStringField(json.furtherCreditName),
		furtherCreditAccount = decodeOptionalStringField(json.furtherCreditAccount),
		correspondentBankAccount = decodeOptionalStringField(json.correspondentBankAccount),
		attention = decodeOptionalStringField(json.attention),
	}
end

function DataTypes.encodeJsonWireInstructionsType(t: WireInstructionsType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = DataTypes.encodeJsonWireInstructionsSubFields(t.valueSubFields)
	end
	return result
end

function DataTypes.decodeJsonWireInstructionsType(json: {[string]: any}): WireInstructionsType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then DataTypes.decodeJsonWireInstructionsSubFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

export type CustomCompoundType = {
	typeId: string,
	valueSubFields: {[string]: FieldValue}?,
	subFieldKeysInOrder: {string},
	label: string?,
}

function DataTypes.makeCustomCompoundType(config: {
	typeId: string?,
	valueSubFields: {[string]: FieldValue}?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): CustomCompoundType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

local STRING_TYPE_IDS: {[string]: boolean} = {
	String = true,
	Ssn = true,
	Ein = true,
	Aba = true,
	Itin = true,
	Swift = true,
	Iban = true,
	Giin = true,
	UsZip = true,
	Email = true,
	PhoneFaxString = true,
	IsoCountryCode = true,
	IsoCurrencyCode = true,
	CountryString = true,
	StateProvince = true,
	City = true,
	NumberAndStreet = true,
	PostalZipCode = true,
	MoneyString = true,
	DateTimeString = true,
	TimeZoneOffset = true
}

local NUMBER_TYPE_IDS: {[string]: boolean} = {
	Integer = true,
	Float = true,
	Percentage = true,
	Year = true
}

local COMPOUND_TYPE_CODECS: {[string]: {decode: string, encode: string}} = {
	Address = { decode = "decodeJsonAddressType", encode = "encodeJsonAddressType" },
	BankAccountInfo = { decode = "decodeJsonBankAccountInfoType", encode = "encodeJsonBankAccountInfoType" },
	BankInfo = { decode = "decodeJsonBankInfoType", encode = "encodeJsonBankInfoType" },
	BaseContact = { decode = "decodeJsonBaseContactType", encode = "encodeJsonBaseContactType" },
	BrokerageAccount = { decode = "decodeJsonBrokerageAccountType", encode = "encodeJsonBrokerageAccountType" },
	BrokerageFirm = { decode = "decodeJsonBrokerageFirmType", encode = "encodeJsonBrokerageFirmType" },
	Country = { decode = "decodeJsonCountryType", encode = "encodeJsonCountryType" },
	DateTime = { decode = "decodeJsonDateTimeType", encode = "encodeJsonDateTimeType" },
	IndividualName = { decode = "decodeJsonIndividualNameType", encode = "encodeJsonIndividualNameType" },
	Money = { decode = "decodeJsonMoneyType", encode = "encodeJsonMoneyType" },
	PhoneFax = { decode = "decodeJsonPhoneFaxType", encode = "encodeJsonPhoneFaxType" },
	ServiceContactPoint = { decode = "decodeJsonServiceContactPointType", encode = "encodeJsonServiceContactPointType" },
	Signatory = { decode = "decodeJsonSignatoryType", encode = "encodeJsonSignatoryType" },
	SubmissionContact = { decode = "decodeJsonSubmissionContactType", encode = "encodeJsonSubmissionContactType" },
	WireInstructions = { decode = "decodeJsonWireInstructionsType", encode = "encodeJsonWireInstructionsType" },
}

local ENUM_TYPE_IDS: {[string]: boolean} = {
	ShareClass = true,
	TransactionType = true,
	SubscriptionStatus = true
}

function DataTypes.decodeFieldValueByTypeId(json: {[string]: any}): FieldValue
	local typeId = json.typeId or ""

	if STRING_TYPE_IDS[typeId] then
		return DataTypes.decodeJsonStringType(json) :: any
	elseif NUMBER_TYPE_IDS[typeId] then
		return DataTypes.decodeJsonNumberType(json) :: any
	elseif typeId == "Boolean" then
		return DataTypes.decodeJsonBooleanType(json) :: any
	elseif ENUM_TYPE_IDS[typeId] then
		return DataTypes.decodeJsonEnumType(json) :: any
	elseif typeId == "MultipleCheckbox" then
		return DataTypes.decodeJsonMultipleCheckboxType(json) :: any
	elseif typeId == "RadioGroup" then
		return DataTypes.decodeJsonRadioGroupType(json) :: any
	elseif typeId == "CustomCompound" or typeId == "CustomCompoundType" then
		return DataTypes.decodeJsonCustomCompoundType(json) :: any
	else
		local codec = COMPOUND_TYPE_CODECS[typeId]
		if codec then
			local decoder = (DataTypes :: any)[codec.decode]
			return decoder(json)
		end
		return json :: any
	end
end

function DataTypes.encodeFieldValueByTypeId(value: FieldValue): {[string]: any}
	local typeId = (value :: any).typeId or ""

	if STRING_TYPE_IDS[typeId] then
		return DataTypes.encodeJsonStringType(value :: any)
	elseif NUMBER_TYPE_IDS[typeId] then
		return DataTypes.encodeJsonNumberType(value :: any)
	elseif typeId == "Boolean" then
		return DataTypes.encodeJsonBooleanType(value :: any)
	elseif ENUM_TYPE_IDS[typeId] then
		return DataTypes.encodeJsonEnumType(value :: any)
	elseif typeId == "MultipleCheckbox" then
		return DataTypes.encodeJsonMultipleCheckboxType(value :: any)
	elseif typeId == "RadioGroup" then
		return DataTypes.encodeJsonRadioGroupType(value :: any)
	elseif typeId == "CustomCompound" or typeId == "CustomCompoundType" then
		return DataTypes.encodeJsonCustomCompoundType(value :: any)
	else
		local codec = COMPOUND_TYPE_CODECS[typeId]
		if codec then
			local encoder = (DataTypes :: any)[codec.encode]
			return encoder(value)
		end
		return value :: any
	end
end

function DataTypes.encodeJsonCustomCompoundType(t: CustomCompoundType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		local encoded: {[string]: any} = {}
		for key, value in t.valueSubFields do
			encoded[key] = DataTypes.encodeFieldValueByTypeId(value)
		end
		result.valueSubFields = encoded
	end
	return result
end

function DataTypes.decodeJsonCustomCompoundType(json: {[string]: any}): CustomCompoundType
	local subFields: {[string]: FieldValue}? = nil
	if json.valueSubFields ~= nil then
		local decoded: {[string]: FieldValue} = {}
		for key, value in json.valueSubFields :: {[string]: any} do
			decoded[key] = DataTypes.decodeFieldValueByTypeId(value)
		end
		subFields = decoded
	end
	return {
		typeId = json.typeId or "",
		valueSubFields = subFields,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

return DataTypes
