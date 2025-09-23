provider "aws" {
  region = "us-east-1" # Replace with your desired AWS region
}

resource "aws_ssm_parameter" "connectors_convenio" {
  name  = "/teste/app/connectors/convenio"
  type  = "String"
  value = "{\"field\":\"convenio\",\"adapter\":\"redis\",\"adapterConfig\":{\"endpoint\":\"localhost:6379\",\"password\":\"\",\"attr\":{\"codigoConvenio\":\"Int\"}},\"keyPattern\":\"CVN_{codigoConvenio}\"}"
}

resource "aws_ssm_parameter" "connectors_limite_operacional" {
  name  = "/teste/app/connectors/limite_operacional"
  type  = "String"
  value = "{\"field\":\"limiteOperacional\",\"adapter\":\"redis\",\"adapterConfig\":{\"endpoint\":\"localhost:6379\",\"password\":\"\",\"attr\":{\"codigoConvenio\":\"Int\"}},\"keyPattern\":\"LMT_{codigoConvenio}\"}"
}

resource "aws_ssm_parameter" "connectors_taxa_funding" {
  name  = "/teste/app/connectors/taxa_funding"
  type  = "String"
  value = "{\"field\":\"taxaFunding\",\"adapter\":\"redis\",\"adapterConfig\":{\"endpoint\":\"localhost:6379\",\"password\":\"\",\"attr\":{\"codigoConvenio\":\"Int\"}},\"keyPattern\":\"FDN_{codigoConvenio}\"}"
}

resource "aws_ssm_parameter" "schema_posibilidade_habilitacao" {
  name  = "/teste/app/schema/types/posibilidade_habilitacao"
  type  = "String"
  value = "{\"name\":\"PossibilidadeHabilitacao\",\"fields\":[{\"name\":\"plataformas\",\"type\":\"List\",\"ofType\":\"String\"},{\"name\":\"modalidades\",\"type\":\"List\",\"ofType\":\"String\"},{\"name\":\"canais\",\"type\":\"List\",\"ofType\":\"String\"}]}"
}

resource "aws_ssm_parameter" "schema_convenio" {
  name  = "/teste/app/schema/types/convenio"
  type  = "String"
  value = "{\"name\":\"Convenio\",\"fields\":[{\"name\":\"codigoConvenio\",\"type\":\"Int\"},{\"name\":\"nomeConvenio\",\"type\":\"String\"},{\"name\":\"apelidoConvenio\",\"type\":\"String\"},{\"name\":\"codigoFq3PessoaJuridica\",\"type\":\"String\"},{\"name\":\"codigoTq3InquilinoPessoaJuridica\",\"type\":\"String\"},{\"name\":\"grupoConsignacao\",\"type\":\"String\"},{\"name\":\"numeroContratoMae\",\"type\":\"String\"},{\"name\":\"instituicaoFinanceira\",\"type\":\"String\"},{\"name\":\"segmento\",\"type\":\"String\"},{\"name\":\"situacaoConvenio\",\"type\":\"String\"},{\"name\":\"possibilidadeHabilitacao\",\"type\":\"Object\",\"ofType\":\"PossibilidadeHabilitacao\"}]}"
}

resource "aws_ssm_parameter" "schema_limite_operacional" {
  name  = "/teste/app/schema/types/limite_operacional"
  type  = "String"
  value = "{\"name\":\"LimiteOperacional\",\"fields\":[{\"name\":\"percentualMaximoMargemConsignavel\",\"type\":\"Float\"},{\"name\":\"percentualMaximoMargemConsignavelRestricao\",\"type\":\"Float\"},{\"name\":\"tipoMargemConsignavelRestricao\",\"type\":\"String\"},{\"name\":\"tipoTaxaMaximaRegulada\",\"type\":\"String\"},{\"name\":\"percentualTaxaMaximaRegulada\",\"type\":\"Float\"},{\"name\":\"quantidadeMinimaPrazo\",\"type\":\"Int\"},{\"name\":\"quantidadeMaximaPrazo\",\"type\":\"Int\"},{\"name\":\"quantidadeMaximaContratacoes\",\"type\":\"Int\"},{\"name\":\"valorMinimoContratacao\",\"type\":\"Float\"},{\"name\":\"valorMaximoContratacao\",\"type\":\"Float\"},{\"name\":\"valorMinimoParcela\",\"type\":\"Float\"},{\"name\":\"idadeMinimaContratacao\",\"type\":\"Int\"},{\"name\":\"idadeMaximaContratacao\",\"type\":\"Int\"},{\"name\":\"valorMinimoTroco\",\"type\":\"Float\"},{\"name\":\"quantidadeMaximaDiasAtraso\",\"type\":\"Int\"},{\"name\":\"percentualMinimoSaldoDevedorQuitado\",\"type\":\"Float\"},{\"name\":\"quantidadeMinimaParcelasPagas\",\"type\":\"Int\"},{\"name\":\"quantidadeMaximaParcelasAtraso\",\"type\":\"Int\"},{\"name\":\"quantidadeMaximaContratosRefinanciadosPorVez\",\"type\":\"Int\"},{\"name\":\"quantidadeMinimaParcelasVencer\",\"type\":\"Int\"}]}"
}

resource "aws_ssm_parameter" "schema_prazo_funding" {
  name  = "/teste/app/schema/types/prazo_funding"
  type  = "String"
  value = "{\"name\":\"PrazoFunding\",\"fields\":[{\"name\":\"prazo\",\"type\":\"Int\"},{\"name\":\"taxaFunding\",\"type\":\"Float\"}]}"
}

resource "aws_ssm_parameter" "schema_taxa_funding" {
  name  = "/teste/app/schema/types/taxa_funding"
  type  = "String"
  value = "{\"name\":\"TaxaFunding\",\"fields\":[{\"name\":\"dataFunding\",\"type\":\"String\"},{\"name\":\"prazos\",\"type\":\"List\",\"ofType\":\"PrazoFunding\"}]}"
}

resource "aws_ssm_parameter" "schema_domicilio_bancario" {
  name  = "/teste/app/schema/types/domicilio_bancario"
  type  = "String"
  value = "{\"name\":\"DomicilioBancario\",\"fields\":[{\"name\":\"codigoAgencia\",\"type\":\"String\"},{\"name\":\"codigoBanco\",\"type\":\"Int\"},{\"name\":\"numeroConta\",\"type\":\"String\"}]}"
}

resource "aws_ssm_parameter" "schema_especie_beneficio" {
  name  = "/teste/app/schema/types/especie_beneficio"
  type  = "String"
  value = "{\"name\":\"EspecieBeneficio\",\"fields\":[{\"name\":\"codigo\",\"type\":\"Int\"},{\"name\":\"descricao\",\"type\":\"String\"}]}"
}

resource "aws_ssm_parameter" "schema_forma_pagamento" {
  name  = "/teste/app/schema/types/forma_pagamento"
  type  = "String"
  value = "{\"name\":\"FormaPagamento\",\"fields\":[{\"name\":\"codigo\",\"type\":\"Int\"},{\"name\":\"descricao\",\"type\":\"String\"}]}"
}

resource "aws_ssm_parameter" "schema_marca_con" {
  name  = "/teste/app/schema/types/marca_con"
  type  = "String"
  value = "{\"name\":\"MarcaCon\",\"fields\":[{\"name\":\"valorComp\",\"type\":\"Float\"},{\"name\":\"valorDae\",\"type\":\"Float\"},{\"name\":\"valorDisC\",\"type\":\"Float\"},{\"name\":\"valorDisE\",\"type\":\"Float\"},{\"name\":\"valorLimC\",\"type\":\"Float\"},{\"name\":\"valorLiq\",\"type\":\"Float\"},{\"name\":\"valorMcb\",\"type\":\"Float\"}]}"
}

resource "aws_ssm_parameter" "schema_origem_mc" {
  name  = "/teste/app/schema/types/origem_mc"
  type  = "String"
  value = "{\"name\":\"OrigemMc\",\"fields\":[{\"name\":\"fonte\",\"type\":\"String\"},{\"name\":\"dataAtualizacao\",\"type\":\"String\"}]}"
}

resource "aws_ssm_parameter" "schema_mensagem_processamento" {
  name  = "/teste/app/schema/types/mensagem_processamento"
  type  = "String"
  value = "{\"name\":\"MensagemProcessamento\",\"fields\":[{\"name\":\"codigos\",\"type\":\"String\"},{\"name\":\"descricao\",\"type\":\"String\"}]}"
}

resource "aws_ssm_parameter" "schema_vinculo" {
  name  = "/teste/app/schema/types/vinculo"
  type  = "String"
  value = "{\"name\":\"Vinculo\",\"fields\":[{\"name\":\"codigoCons\",\"type\":\"String\"},{\"name\":\"codigoMbf\",\"type\":\"String\"},{\"name\":\"domicilioBancario\",\"type\":\"Object\",\"ofType\":\"DomicilioBancario\"},{\"name\":\"dataBb\",\"type\":\"String\"},{\"name\":\"dataCm\",\"type\":\"String\"},{\"name\":\"dataDb\",\"type\":\"String\"},{\"name\":\"dataExtc\",\"type\":\"String\"},{\"name\":\"dataNc\",\"type\":\"String\"},{\"name\":\"dataTcB\",\"type\":\"String\"},{\"name\":\"dataTrl\",\"type\":\"String\"},{\"name\":\"dataUppc\",\"type\":\"String\"},{\"name\":\"especieBeneficio\",\"type\":\"Object\",\"ofType\":\"EspecieBeneficio\"},{\"name\":\"formaPagamento\",\"type\":\"Object\",\"ofType\":\"FormaPagamento\"},{\"name\":\"codigoIdentificacaoInquilino\",\"type\":\"String\"},{\"name\":\"codigoIdentificacaoRt\",\"type\":\"String\"},{\"name\":\"indicadorEb\",\"type\":\"Boolean\"},{\"name\":\"indicadorEe\",\"type\":\"Boolean\"},{\"name\":\"indicadorOr\",\"type\":\"Boolean\"},{\"name\":\"indicadorPa\",\"type\":\"Object\",\"ofType\":\"FormaPagamento\"},{\"name\":\"indicadorPep\",\"type\":\"Object\",\"ofType\":\"FormaPagamento\"},{\"name\":\"indicadorPro\",\"type\":\"Boolean\"},{\"name\":\"indicadorRt\",\"type\":\"Boolean\"},{\"name\":\"indicadorSben\",\"type\":\"Object\",\"ofType\":\"FormaPagamento\"},{\"name\":\"indicadorCj\",\"type\":\"Boolean\"},{\"name\":\"marcaCon\",\"type\":\"Object\",\"ofType\":\"MarcaCon\"},{\"name\":\"nomeDc\",\"type\":\"String\"},{\"name\":\"origemMc\",\"type\":\"List\",\"ofType\":\"OrigemMc\"},{\"name\":\"quantidadeEmAs\",\"type\":\"Int\"},{\"name\":\"quantidadeEma\",\"type\":\"Int\"},{\"name\":\"quantidadeEmp\",\"type\":\"Int\"},{\"name\":\"quantidadeEmr\",\"type\":\"Int\"},{\"name\":\"quantidadeEms\",\"type\":\"Int\"},{\"name\":\"quantidadePl\",\"type\":\"Int\"},{\"name\":\"siglaUfp\",\"type\":\"String\"},{\"name\":\"tipoBloq\",\"type\":\"Object\",\"ofType\":\"FormaPagamento\"},{\"name\":\"tipoVinCol\",\"type\":\"String\"},{\"name\":\"valorCbene\",\"type\":\"Float\"},{\"name\":\"valorLib\",\"type\":\"Float\"},{\"name\":\"mensagemProcessamento\",\"type\":\"Object\",\"ofType\":\"MensagemProcessamento\"}]}"
}

resource "aws_ssm_parameter" "schema_colaborador_convenio" {
  name  = "/teste/app/schema/types/colaborador_convenio"
  type  = "String"
  value = "{\"name\":\"ColaboradorConvenio\",\"fields\":[{\"name\":\"codigoIdentificacaoPessoa\",\"type\":\"String\"},{\"name\":\"codigoConvenioContrato\",\"type\":\"String\"},{\"name\":\"vinculos\",\"type\":\"List\",\"ofType\":\"Vinculo\"}]}"
}

resource "aws_ssm_parameter" "schema_combined_data" {
  name  = "/teste/app/schema/types/combined_data"
  type  = "String"
  value = "{\"name\":\"CombinedData\",\"fields\":[{\"name\":\"convenio\",\"type\":\"Object\",\"ofType\":\"Convenio\"},{\"name\":\"limiteOperacional\",\"type\":\"Object\",\"ofType\":\"LimiteOperacional\"},{\"name\":\"taxaFunding\",\"type\":\"Object\",\"ofType\":\"TaxaFunding\"},{\"name\":\"colaboradorConvenio\",\"type\":\"Object\",\"ofType\":\"ColaboradorConvenio\"}]}"
}

resource "aws_ssm_parameter" "schema_query" {
  name  = "/teste/app/schema/query"
  type  = "String"
  value = "{\"name\":\"Query\",\"fields\":[{\"name\":\"dataSources\",\"type\":\"Object\",\"ofType\":\"CombinedData\",\"args\":[{\"name\":\"codigoConvenio\",\"type\":\"NonNull\",\"ofType\":\"Int\"}]}]}"
}
