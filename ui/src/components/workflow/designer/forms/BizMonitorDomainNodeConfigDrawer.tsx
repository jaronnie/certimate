import { useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { Form } from "antd";

import { NodeConfigDrawer } from "./_shared";
import BizMonitorDomainNodeConfigForm from "./BizMonitorDomainNodeConfigForm";
import { NodeType } from "../nodes/typings";

export interface BizMonitorDomainNodeConfigDrawerProps {
  afterClose?: () => void;
  loading?: boolean;
  node: FlowNodeEntity;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

const BizMonitorDomainNodeConfigDrawer = ({ node, ...props }: BizMonitorDomainNodeConfigDrawerProps) => {
  if (node.flowNodeType !== NodeType.BizMonitorDomain) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizMonitorDomain}`);
  }

  const { i18n } = useTranslation();

  const [formInst] = Form.useForm();

  return (
    <NodeConfigDrawer
      anchor={{
        items: BizMonitorDomainNodeConfigForm.getAnchorItems({ i18n }),
      }}
      form={formInst}
      node={node}
      {...props}
    >
      <BizMonitorDomainNodeConfigForm form={formInst} node={node} />
    </NodeConfigDrawer>
  );
};

export default BizMonitorDomainNodeConfigDrawer;
