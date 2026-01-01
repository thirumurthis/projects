package com.spring.grpc.handler;

import com.spring.grpc.dto.OrderStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface OrderStatusService extends JpaRepository<OrderStatus,Long> {

    List<OrderStatus> findByOrderId(long orderId);
    List<OrderStatus> findByUserName(String userName);

    List<OrderStatus> findByUserNameOrOrderId(String userName, long orderId);

}
