// SPDX-License-Identifier: Apache-2.0

/*
 * Copyright 2022, Specular contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

pragma solidity ^0.8.0;

/**
 * @notice Data availability interface to rollup contracts.
 */
interface IDAProvider {
    /**
     * @notice Gets inbox size (total number of messages stored).
     */
    function getInboxSize() external view returns (uint256);
    /**
     * Verifies proof of inclusion of a transaction by the data availability provider.
     * If verification fails, the function reverts.
     * @param encodedTx RLP-encoded transaction.
     * @param proof DA-specific membership proof.
     */
    function verifyTxInclusion(bytes memory encodedTx, bytes calldata proof) external view;
}
