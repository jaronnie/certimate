import { useMemo } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { type AnchorProps, Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";
import { type WorkflowNodeConfigForBizMonitorDomain, defaultNodeConfigForBizMonitorDomain } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { isDomain } from "@/utils/validator";

import { NodeFormContextProvider } from "./_context";
import { NodeType } from "../nodes/typings";

export interface BizMonitorDomainNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const BizMonitorDomainNodeConfigForm = ({ node, ...props }: BizMonitorDomainNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizMonitorDomain) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizMonitorDomain}`);
  }

  const { i18n, t } = useTranslation();

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForBizMonitorDomain | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizMonitorDomainConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item>
            <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.monitor_domain.form.guide") }}></span>} />
          </Form.Item>

          <Form.Item name="domain" label={t("workflow_node.monitor_domain.form.domain.label")} rules={[formRule]}>
            <Input placeholder={t("workflow_node.monitor_domain.form.domain.placeholder")} />
          </Form.Item>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters"].map((key) => ({
    key: key,
    title: t(`workflow_node.monitor_domain.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return defaultNodeConfigForBizMonitorDomain();
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    domain: z.string().refine((v) => isDomain(v), t("common.errmsg.domain_invalid")),
  });
};

const _default = Object.assign(BizMonitorDomainNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
